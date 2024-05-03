package utils

import (
	"context"
	"cosmossdk.io/log"
	"encoding/hex"
	"github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/pkg/errors"
	"time"
)

type RPCHelper struct {
	log    log.Logger
	cliCtx client.Context
}

// NewRPCHelper creates a simple interface to interact with the RPC server.
// Underneath, it doesn't care which node it is connected to.
// - Assumes that cdc and interfaceRegistry already registered all the necessary types. (ophost/opchild inclusively)
// - Assumes that txConfig is already set up.
func NewRPCHelper(
	log log.Logger,
	rpcAddr string,
	chainId string,
	cdc codec.Codec,
	interfaceRegistry types.InterfaceRegistry,
	txConfig client.TxConfig,
) (*RPCHelper, error) {
	httpCli, err := http.New(rpcAddr, "/websocket")
	if err != nil {
		return nil, err
	}

	return &RPCHelper{
		log: log,
		cliCtx: client.Context{}.
			WithClient(httpCli).
			WithChainID(chainId).
			WithCodec(cdc).
			WithInterfaceRegistry(interfaceRegistry).
			WithTxConfig(txConfig),
	}, nil
}

// GetNonce returns the account information for the given address
func (r *RPCHelper) GetNonce(address string) (client.Account, error) {
	addr, _ := sdk.AccAddressFromBech32(address)
	ar := authtypes.AccountRetriever{}
	return ar.GetAccount(
		r.cliCtx,
		addr,
	)
}

// BroadcastTxAndWait broadcasts a transaction and waits until it is included in a block.
func (r *RPCHelper) BroadcastTxAndWait(
	senderAddress string,
	mnemonic string,
	gasLimit uint64,
	fee sdk.Coins,
	msgs ...sdk.Msg,
) (*coretypes.ResultTx, error) {
	r.log.Info("building tx...",
		"fee", fee.String(),
		"msgs-len", len(msgs),
	)

	acc, err := r.GetNonce(senderAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get nonce for %s", senderAddress)
	}

	signedTx, err := SignTxOffline(
		&r.cliCtx,
		mnemonic,
		gasLimit,
		acc.GetAccountNumber(),
		acc.GetSequence(),
		fee,
		msgs...,
	)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to sign tx")
	}

	// encode
	txBytes, err := r.cliCtx.TxConfig.TxEncoder()(signedTx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode transaction")
	}

	json, err := r.cliCtx.TxConfig.TxJSONEncoder()(signedTx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to encode transaction to json")
	}
	r.log.Info("built transaction", "tx-bytes", string(json[:100]))

	// broadcast
	txResponse, err := r.cliCtx.BroadcastTxSync(txBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to broadcast transaction")
	}

	// handle if code is not 0
	if txResponse.Code != 0 {
		return nil, errors.Errorf("tx failed with code %d, raw_log: %s", txResponse.Code, txResponse.RawLog)
	}

	r.log.Info("broadcasted transaction",
		"tx-response", txResponse.TxHash,
		"tx-height", txResponse.Height,
	)

	// poll until tx is found
	hashBz, _ := hex.DecodeString(txResponse.TxHash)

	r.log.Info("waiting until tx is found...",
		"tx-hash", txResponse.TxHash,
	)

	for retry := 0; retry < 10; retry++ {
		txResult, _ := r.cliCtx.Client.Tx(
			context.Background(),
			hashBz,
			false,
		)

		if txResult != nil {
			r.log.Info("found transaction",
				"tx-hash", txResponse.TxHash,
				"tx-height", txResponse.Height,
				"log", txResponse.RawLog,
			)
			return txResult, nil
		}

		time.Sleep(3 * time.Second)
	}

	// shouldn't reach here
	return nil, errors.New("failed to find transaction after 10 tries")
}
