package utils

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/pkg/errors"

	"github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

type RPCHelper struct {
	isL1   bool
	log    log.Logger
	cliCtx client.Context
}

// NewRPCHelper creates a simple interface to interact with the RPC server.
// Underneath, it doesn't care which node it is connected to.
// - Assumes that cdc and interfaceRegistry already registered all the necessary types. (ophost/opchild inclusively)
// - Assumes that txConfig is already set up.
func NewRPCHelper(
	isL1 bool,
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

	cdc.InterfaceRegistry().SigningContext().AddressCodec()
	return &RPCHelper{
		isL1: isL1,
		log:  log,
		cliCtx: client.Context{}.
			WithClient(httpCli).
			WithChainID(chainId).
			WithCodec(cdc).
			WithInterfaceRegistry(interfaceRegistry).
			WithTxConfig(txConfig),
	}, nil
}

// GetStatus returns the status of the chain
func (r *RPCHelper) GetStatus() (*coretypes.ResultStatus, error) {
	return r.cliCtx.Client.Status(context.Background())
}

// getNonce returns the account information for the given address
func (r *RPCHelper) getNonce(addrStr string) (client.Account, error) {
	var ac address.Codec
	if r.isL1 {
		ac = L1AddressCodec()
	} else {
		ac = L2AddressCodec()
	}

	addr, err := ac.StringToBytes(addrStr)
	if err != nil {
		return nil, err
	}

	ar := authtypes.AccountRetriever{}
	return ar.GetAccount(
		r.cliCtx,
		addr,
	)
}

func (r *RPCHelper) GetBridgeInfo(bridgeId uint64) (*ophosttypes.QueryBridgeResponse, error) {
	bz, err := r.cliCtx.Codec.Marshal(&ophosttypes.QueryBridgeRequest{
		BridgeId: bridgeId,
	})
	if err != nil {
		return nil, err
	}

	res, err := r.cliCtx.Client.ABCIQuery(
		context.Background(),
		"/opinit.ophost.v1.Query/Bridge",
		bz,
	)
	if err != nil {
		return nil, err
	}
	if res.Response.Code != 0 {
		return nil, errors.Errorf("failed to query bridge info: %s", res.Response.Log)
	}

	var bridgeInfoRes ophosttypes.QueryBridgeResponse
	err = r.cliCtx.Codec.Unmarshal(res.Response.Value, &bridgeInfoRes)
	if err != nil {
		return nil, err
	}

	return &bridgeInfoRes, nil
}

// CreateAndSignTx creates and signs a transaction
func (r *RPCHelper) CreateAndSignTx(
	senderAddress string,
	mnemonic string,
	gasLimit uint64,
	fee sdk.Coins,
	msgs ...sdk.Msg,
) ([]byte, error) {
	r.log.Info("building tx...",
		"fee", fee.String(),
		"msgs-len", len(msgs),
	)

	if r.isL1 {
		cleanup := HackBech32Prefix("init")
		defer cleanup()
	}

	acc, err := r.getNonce(senderAddress)
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

	return txBytes, nil
}

// BroadcastTxAndWait broadcasts a transaction and waits until it is included in a block.
func (r *RPCHelper) BroadcastTxAndWait(
	senderAddress string,
	mnemonic string,
	gasLimit uint64,
	fee sdk.Coins,
	msgs ...sdk.Msg,
) (*coretypes.ResultTx, error) {
	txBytes, err := r.CreateAndSignTx(senderAddress, mnemonic, gasLimit, fee, msgs...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create and sign transaction")
	}

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

	for range 10 {
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
