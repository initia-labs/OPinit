package l2connect

import (
	"bytes"
	"fmt"

	cometabci "github.com/cometbft/cometbft/abci/types"
	cryptoenc "github.com/cometbft/cometbft/crypto/encoding"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	protoio "github.com/cosmos/gogoproto/io"
	"github.com/cosmos/gogoproto/proto"
)

// ValidateVoteExtensions defines a helper function for verifying vote extension
// signatures that may be passed or manually injected into a block proposal from
// a proposer in PrepareProposal. It returns an error if any signature is invalid
// or if unexpected vote extensions and/or signatures are found or less than 2/3
// power is received.
func ValidateVoteExtensions(
	ctx sdk.Context,
	valStore ValidatorStore,
	height int64,
	chainID string,
	extCommit cometabci.ExtendedCommitInfo,
) error {
	marshalDelimitedFn := func(msg proto.Message) ([]byte, error) {
		var buf bytes.Buffer
		if err := protoio.NewDelimitedWriter(&buf).WriteMsg(msg); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}

	var (
		// Total voting power of all validators in current validator store.
		totalVP int64
		// Total voting power of all validators that submitted valid vote extensions.
		sumVP int64
	)

	totalBondedTokens, err := valStore.TotalBondedTokens(ctx)
	if err != nil {
		return err
	}
	totalVP = totalBondedTokens.Int64()

	for _, vote := range extCommit.Votes {
		valConsAddr := sdk.ConsAddress(vote.Validator.Address)
		power, err := valStore.GetPowerByConsAddr(ctx, valConsAddr)
		if err != nil {
			// return fmt.Errorf("failed to get validator %X power: %w", valConsAddr, err)

			// use only current validator set, so ignore if the validator of the vote is not in the set.
			continue
		}

		if vote.BlockIdFlag == cmtproto.BlockIDFlagCommit && len(vote.ExtensionSignature) == 0 {
			return fmt.Errorf("vote extension signature is missing; validator addr %s",
				vote.Validator.String(),
			)
		}
		if vote.BlockIdFlag != cmtproto.BlockIDFlagCommit && len(vote.VoteExtension) != 0 {
			return fmt.Errorf("non-commit vote extension present; validator addr %s",
				vote.Validator.String(),
			)
		}
		if vote.BlockIdFlag != cmtproto.BlockIDFlagCommit && len(vote.ExtensionSignature) != 0 {
			return fmt.Errorf("non-commit vote extension signature present; validator addr %s",
				vote.Validator.String(),
			)
		}

		// Only check + include power if the vote is a commit vote. There must be super-majority, otherwise the
		// previous block (the block vote is for) could not have been committed.
		if vote.BlockIdFlag != cmtproto.BlockIDFlagCommit {
			continue
		}

		pubKeyProto, err := valStore.GetPubKeyByConsAddr(ctx, valConsAddr)
		if err != nil {
			return fmt.Errorf("failed to get validator %X public key: %w", valConsAddr, err)
		}

		cmtPubKey, err := cryptoenc.PubKeyFromProto(pubKeyProto)
		if err != nil {
			return fmt.Errorf("failed to convert validator %X public key: %w", valConsAddr, err)
		}

		cve := cmtproto.CanonicalVoteExtension{
			ChainId:   chainID,
			Height:    height,
			Round:     int64(extCommit.Round),
			Extension: vote.VoteExtension,
		}

		extSignBytes, err := marshalDelimitedFn(&cve)
		if err != nil {
			return fmt.Errorf("failed to encode CanonicalVoteExtension: %w", err)
		}

		if !cmtPubKey.VerifySignature(extSignBytes, vote.ExtensionSignature) {
			return fmt.Errorf("failed to verify validator %X vote extension signature", valConsAddr)
		}

		sumVP += power.Int64()
	}

	// This check is probably unnecessary, but better safe than sorry.
	if totalVP <= 0 {
		return fmt.Errorf("total voting power must be positive, got: %d", totalVP)
	}

	// If the sum of the voting power has not reached (2/3 + 1) we need to error.
	if requiredVP := ((totalVP * 2) / 3) + 1; sumVP < requiredVP {
		return fmt.Errorf(
			"insufficient cumulative voting power received to verify vote extensions; got: %d, expected: >=%d",
			sumVP, requiredVP,
		)
	}

	return nil
}
