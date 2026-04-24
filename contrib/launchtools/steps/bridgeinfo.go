package steps

import (
	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v10/modules/core/exported"
	tmclient "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"

	"github.com/initia-labs/OPinit/contrib/launchtools"
	opchildtypes "github.com/initia-labs/OPinit/x/opchild/types"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
)

var _ launchtools.LauncherStepFuncFactory[*launchtools.Config] = SetBridgeInfo

// SetBridgeInfo creates OP bridge between OPChild and OPHost
func SetBridgeInfo(
	config *launchtools.Config,
) launchtools.LauncherStepFunc {
	return func(ctx launchtools.Launcher) error {
		ctx.Logger().Info("SetBridgeInfo")

		if ctx.GetBridgeId() == nil {
			return errors.New("bridge ID not initialized")
		}

		if ctx.GetRPCHelperL1() == nil {
			return errors.New("RPC helper for L1 not initialized")
		}

		if ctx.GetRelayer() == nil {
			return errors.New("relayer not initialized")
		}

		// scan all states from IBC - get all established states.
		var l1ClientID string
		ctx.App().GetIBCKeeper().ClientKeeper.IterateClientStates(
			sdk.UnwrapSDKContext(ctx.QueryContext()), nil,
			func(clientID string, clientState exported.ClientState) bool {
				tcli, ok := clientState.(*tmclient.ClientState)
				if !ok {
					ctx.Logger().Debug("skipping non-tendermint client", "client-type", clientState.String())
					return false
				}
				if tcli.ChainId == config.L1Config.ChainID {
					ctx.Logger().Info("found L1 tendermint client", "client-id", clientID)
					l1ClientID = clientID
					return true
				}
				return false
			},
		)

		// if L1 tendermint client is never found, return an error
		if l1ClientID == "" {
			return errors.New("failed to find L1 tendermint client")
		}

		bridgeId := *ctx.GetBridgeId()
		bridgeInfo, err := ctx.GetRPCHelperL1().GetBridgeInfo(bridgeId)
		if err != nil {
			return errors.Wrapf(err, "failed to get bridge info from L1")
		}

		// create SetBridgeInfo message
		setBridgeInfoMessage := setBridgeInfo(
			config.SystemKeys.BridgeExecutor.L2Address,
			bridgeId,
			bridgeInfo.BridgeAddr,
			config.L1Config.ChainID,
			l1ClientID,
			bridgeInfo.BridgeConfig,
		)
		// send MsgSetBridgeInfo to host (L1)
		txRes, err := ctx.GetRPCHelperL2().BroadcastTxAndWait(
			config.SystemKeys.BridgeExecutor.L2Address,
			config.SystemKeys.BridgeExecutor.Mnemonic,
			200000,
			sdk.NewCoins(),
			setBridgeInfoMessage,
		)
		if err != nil {
			return errors.Wrap(err, "failed to broadcast tx")
		}

		// if transaction failed, return error
		if txRes.TxResult.Code != 0 {
			ctx.Logger().Error("tx failed", "code", txRes.TxResult.Code, "log", txRes.TxResult.Log)
			return errors.Errorf("tx failed with code %d", txRes.TxResult.Code)
		}

		// update client state
		err = ctx.GetRelayer().UpdateClients()
		if err != nil {
			return errors.Wrap(err, "failed to update clients")
		}

		return nil
	}
}

func setBridgeInfo(
	executorAddress string,
	bridgeId uint64,
	bridgeAddr string,
	l1ChainId string,
	l1ClientId string,
	bridgeConfig ophosttypes.BridgeConfig,
) *opchildtypes.MsgSetBridgeInfo {
	return opchildtypes.NewMsgSetBridgeInfo(
		executorAddress,
		opchildtypes.BridgeInfo{
			BridgeId:     bridgeId,
			BridgeAddr:   bridgeAddr,
			L1ChainId:    l1ChainId,
			L1ClientId:   l1ClientId,
			BridgeConfig: bridgeConfig,
		},
	)
}
