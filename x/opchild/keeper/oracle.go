package keeper

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	ics23 "github.com/cosmos/ics23/go"
	oracletypes "github.com/skip-mev/connect/v2/x/oracle/types"

	connectaggregator "github.com/skip-mev/connect/v2/abci/strategies/aggregator"
	connectcodec "github.com/skip-mev/connect/v2/abci/strategies/codec"
	"github.com/skip-mev/connect/v2/abci/strategies/currencypair"
	"github.com/skip-mev/connect/v2/pkg/math/voteweighted"
	connecttypes "github.com/skip-mev/connect/v2/pkg/types"

	"github.com/initia-labs/OPinit/x/opchild/l2connect"
	"github.com/initia-labs/OPinit/x/opchild/types"
)

type L2OracleHandler struct {
	*Keeper

	oracleKeeper        types.OracleKeeper
	extendedCommitCodec connectcodec.ExtendedCommitCodec
	veCodec             connectcodec.VoteExtensionCodec
}

func NewL2OracleHandler(
	k *Keeper,
	oracleKeeper types.OracleKeeper,
	logger log.Logger,
) *L2OracleHandler {
	return &L2OracleHandler{
		Keeper:       k,
		oracleKeeper: oracleKeeper,
		extendedCommitCodec: connectcodec.NewCompressionExtendedCommitCodec(
			connectcodec.NewDefaultExtendedCommitCodec(),
			connectcodec.NewZStdCompressor(),
		),
		veCodec: connectcodec.NewCompressionVoteExtensionCodec(
			connectcodec.NewDefaultVoteExtensionCodec(),
			connectcodec.NewZLibCompressor(),
		),
	}
}

func (k L2OracleHandler) UpdateOracle(ctx context.Context, height uint64, extCommitBz []byte) error {
	hostStoreLastHeight, err := k.HostValidatorStore.GetLastHeight(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.ErrOracleValidatorsNotRegistered
		}
		return err
	}

	h := int64(height) //nolint:gosec
	if hostStoreLastHeight-5 > h {
		return types.ErrInvalidOracleHeight
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	hostChainID, err := k.L1ChainId(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.ErrBridgeInfoNotExists
		}
		return err
	}

	extendedCommitInfo, err := k.extendedCommitCodec.Decode(extCommitBz)
	if err != nil {
		return err
	}

	extendedVotes, err := l2connect.ValidateVoteExtensions(sdkCtx, k.HostValidatorStore, h-1, hostChainID, extendedCommitInfo)
	if err != nil {
		return err
	}

	votes, err := l2connect.GetOracleVotes(k.veCodec, extendedVotes)
	if err != nil {
		return err
	}

	// create a new vote aggregator for each update
	voteAggregator := connectaggregator.NewDefaultVoteAggregator(
		k.Logger(ctx),
		voteweighted.MedianFromContext(
			k.Logger(ctx),
			k.HostValidatorStore,
			voteweighted.DefaultPowerThreshold,
		),
		currencypair.NewHashCurrencyPairStrategy(k.oracleKeeper),
	)
	prices, err := voteAggregator.AggregateOracleVotes(sdkCtx, votes)
	if err != nil {
		return err
	}

	tsCp, err := connecttypes.CurrencyPairFromString(l2connect.ReservedCPTimestamp)
	if err != nil {
		return err
	}

	// if there is no timestamp price, skip the price update
	if _, ok := prices[tsCp]; !ok {
		return types.ErrOracleTimestampNotExists
	}

	updatedTime := time.Unix(0, prices[tsCp].Int64())
	err = l2connect.WritePrices(sdkCtx, k.oracleKeeper, updatedTime, prices)
	if err != nil {
		return err
	}

	return nil
}

// HandleOracleDataPacket handles the oracle data relayed from L1.
// This verifies the oracle state proof and applies the price update to L2's oracle module.
func (k Keeper) HandleOracleDataPacket(
	ctx context.Context,
	oracleData types.OracleData,
) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	bridgeInfo, err := k.BridgeInfo.Get(ctx)
	if err != nil {
		return types.ErrBridgeInfoNotExists
	}

	if oracleData.BridgeId != bridgeInfo.BridgeId {
		return types.ErrInvalidBridgeInfo
	}

	if !bridgeInfo.BridgeConfig.OracleEnabled {
		return types.ErrOracleDisabled
	}

	if err := k.verifyOracleDataProof(ctx, oracleData, bridgeInfo); err != nil {
		return errorsmod.Wrap(err, "oracle state proof verification failed")
	}

	if err := k.processOraclePriceUpdate(ctx, oracleData); err != nil {
		return errorsmod.Wrap(err, "failed to process oracle price")
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeOracleDataRelay,
			sdk.NewAttribute(types.AttributeKeyBridgeId, fmt.Sprintf("%d", oracleData.BridgeId)),
			sdk.NewAttribute(types.AttributeKeyL1BlockHeight, fmt.Sprintf("%d", oracleData.L1BlockHeight)),
			sdk.NewAttribute(types.AttributeKeyCurrencyPair, oracleData.CurrencyPair),
			sdk.NewAttribute(types.AttributeKeyPrice, oracleData.Price),
		),
	)

	return nil
}

// verifyOracleDataProof verifies the oracle state proof from L1.
// This verifies that the oracle price data existed in L1's oracle module state
// at the specified height using a Merkle proof against L1's state root.
func (k Keeper) verifyOracleDataProof(
	ctx context.Context,
	data types.OracleData,
	bridgeInfo types.BridgeInfo,
) error {
	k.ensureIBCKeepersSet()
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// construct the state path for the oracle data in the oracle module
	cpBytes := []byte(data.CurrencyPair)
	stateKey := append([]byte{0x00}, cpBytes...)

	priceInt, ok := math.NewIntFromString(data.Price)
	if !ok {
		return sdkerrors.ErrInvalidRequest.Wrapf("invalid price format: %s", data.Price)
	}

	sec := data.L1BlockTime / 1e9
	nsec := data.L1BlockTime % 1e9
	quotePrice := oracletypes.QuotePrice{
		Price:          priceInt,
		BlockTimestamp: time.Unix(sec, nsec),
		BlockHeight:    data.L1BlockHeight,
	}
	expectedCurrencyPairState := oracletypes.CurrencyPairState{
		Price: &quotePrice,
		Nonce: data.Nonce,
		Id:    data.CurrencyPairId,
	}
	expectedValue := k.cdc.MustMarshal(&expectedCurrencyPairState)

	proofHeight := clienttypes.NewHeight(data.ProofHeight.RevisionNumber, data.ProofHeight.RevisionHeight)

	// convert ProofOps to ICS-23 MerkleProof format
	merkleProof, err := convertProofOpsToMerkleProof(data.Proof)
	if err != nil {
		return errorsmod.Wrap(err, "failed to convert proof to merkle proof")
	}
	merkleProofBz, err := k.cdc.Marshal(&merkleProof)
	if err != nil {
		return errorsmod.Wrap(err, "failed to marshal merkle proof")
	}

	// path for verification: oracle store key + state key
	// the merkle path should match how the data is stored in the iavl tree
	merklePath := commitmenttypes.NewMerklePath(oracletypes.StoreKey, string(stateKey))

	clientState, found := k.ibcClientKeeper.GetClientState(sdkCtx, bridgeInfo.L1ClientId)
	if !found {
		return errors.New("L1 IBC client state not found")
	}
	clientStore := k.ibcClientKeeper.ClientStore(sdkCtx, bridgeInfo.L1ClientId)

	// verify the membership proof using the ibc light client
	if err := clientState.VerifyMembership(
		sdkCtx,
		clientStore,
		k.Codec(),
		proofHeight,
		0,
		0,
		merkleProofBz,
		merklePath,
		expectedValue,
	); err != nil {
		k.Logger(ctx).Error("oracle state proof verification failed",
			"error", err.Error(),
			"path", merklePath.String(),
		)
		return errorsmod.Wrap(err, "oracle state proof verification failed")
	}

	return nil
}

// processOraclePriceUpdate processes a single currency pair price update from L1.
func (k Keeper) processOraclePriceUpdate(ctx context.Context, oracleData types.OracleData) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	cp, err := connecttypes.CurrencyPairFromString(oracleData.CurrencyPair)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid currency pair format: %s", oracleData.CurrencyPair)
	}

	// check if incoming oracle data is newer than existing using timestamp comparison
	l1Time := time.Unix(0, oracleData.L1BlockTime)
	existingPrice, err := k.l2OracleHandler.oracleKeeper.GetPriceForCurrencyPair(ctx, cp)
	if err == nil && !l1Time.After(existingPrice.BlockTimestamp) {
		return types.ErrInvalidOracleTimestamp
	}

	priceInt, ok := math.NewIntFromString(oracleData.Price)
	if !ok {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid price format")
	}

	// we store l1 timestamp for staleness checks, l2 block height for reference
	qp := oracletypes.QuotePrice{
		Price:          priceInt,
		BlockTimestamp: l1Time,
		BlockHeight:    uint64(sdkCtx.BlockHeight()), //nolint:gosec
	}

	if err := k.l2OracleHandler.oracleKeeper.SetPriceForCurrencyPair(ctx, cp, qp); err != nil {
		return errorsmod.Wrap(err, "failed to set price for currency pair")
	}

	return nil
}

// convertProofOpsToMerkleProof converts ABCI ProofOps to ICS-23 MerkleProof format.
// The ABCI query returns ProofOps where each op's data field contains an ICS-23 CommitmentProof.
func convertProofOpsToMerkleProof(proofBz []byte) (commitmenttypes.MerkleProof, error) {
	var proofOps crypto.ProofOps
	if err := proofOps.Unmarshal(proofBz); err != nil {
		return commitmenttypes.MerkleProof{}, errorsmod.Wrap(err, "failed to unmarshal proof ops")
	}

	var proofs []*ics23.CommitmentProof
	for _, op := range proofOps.Ops {
		var commitmentProof ics23.CommitmentProof
		if err := commitmentProof.Unmarshal(op.Data); err != nil {
			return commitmenttypes.MerkleProof{}, errorsmod.Wrapf(err, "failed to unmarshal commitment proof for op type %s", op.Type)
		}
		proofs = append(proofs, &commitmentProof)
	}

	return commitmenttypes.MerkleProof{Proofs: proofs}, nil
}
