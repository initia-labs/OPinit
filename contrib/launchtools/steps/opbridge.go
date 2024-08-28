package steps

import (
	"encoding/json"
	"strconv"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/initia-labs/OPinit/contrib/launchtools"
	"github.com/initia-labs/OPinit/contrib/launchtools/utils"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
	ophosthooktypes "github.com/initia-labs/OPinit/x/ophost/types/hook"
	"github.com/pkg/errors"
)

var _ launchtools.LauncherStepFuncFactory[*launchtools.Config] = InitializeOpBridge

const BridgeArtifactName = "BRIDGE_ID"

// InitializeOpBridge creates OP bridge between OPChild and OPHost
func InitializeOpBridge(
	config *launchtools.Config,
) launchtools.LauncherStepFunc {
	return func(ctx launchtools.Launcher) error {
		ctx.Logger().Info("initializing OpBridge")

		// scan all channels. This should give all IBC channels established before this step
		// - assumes channels aren't longer than the default pagination limit
		channels, err := ctx.App().GetIBCKeeper().Channels(
			ctx.QueryContext(),
			&ibctypes.QueryChannelsRequest{}, // assume there aren't many channels already open
		)
		if err != nil {
			return errors.Wrap(err, "failed to query client states")
		}

		ctx.Logger().Info("found channels", "channels", channels.Channels)

		// create OpBridgeMessage
		createOpBridgeMessage, err := createOpBridge(
			channels.Channels,
			config.SystemKeys.BridgeExecutor.L1Address,
			config.SystemKeys.Challenger.L1Address,
			config.SystemKeys.OutputSubmitter.L1Address,
			config.SystemKeys.BatchSubmitter.L1Address,
			config.OpBridge.BatchSubmissionTarget,
			*config.OpBridge.OutputSubmissionInterval,
			*config.OpBridge.OutputFinalizationPeriod,
			config.OpBridge.OutputSubmissionStartHeight,
			*config.OpBridge.EnableOracle,
		)

		ctx.Logger().Info("creating op bridge...", "message", createOpBridgeMessage.String())

		if err != nil {
			return errors.Wrap(err, "failed to create OpBridgeMessage")
		}

		ctx.Logger().Info("broadcasting tx to L1...",
			"from-address", config.SystemKeys.BridgeExecutor.L1Address,
		)

		// already validated in config.go
		gasPrices, _ := sdk.ParseDecCoins(config.L1Config.GasPrices)
		gasFees, _ := gasPrices.MulDec(math.LegacyNewDecFromInt(math.NewInt(200000))).TruncateDecimal()

		// send createOpBridgeMessage to host (L1)
		res, err := ctx.GetRPCHelperL1().BroadcastTxAndWait(
			config.SystemKeys.BridgeExecutor.L1Address,
			config.SystemKeys.BridgeExecutor.Mnemonic,
			200000,
			gasFees,
			createOpBridgeMessage,
		)
		if err != nil {
			return errors.Wrap(err, "failed to broadcast tx")
		}

		// if transaction failed, return error
		if res.TxResult.Code != 0 {
			ctx.Logger().Error("tx failed", "code", res.TxResult.Code, "log", res.TxResult.Log)
			return errors.Errorf("tx failed with code %d", res.TxResult.Code)
		}

		// otherwise find OpBridgeID in tx events
		opBridgeId, found := utils.FindTxEventsByKey(OpBridgeIDKey, res.TxResult.Events)
		if !found {
			return errors.Errorf("failed to find OpBridgeID")
		}

		ctx.Logger().Info("opbridge created", "op-bridge-id", opBridgeId)

		bridgeId, err := strconv.ParseUint(opBridgeId, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "failed to parse OpBridgeID %s", opBridgeId)
		}
		ctx.SetBridgeId(bridgeId)

		// otherwise write OpBridgeID to file and return
		return ctx.WriteOutput(BridgeArtifactName, opBridgeId)
	}
}

func createOpBridge(
	identifiedChannels []*ibctypes.IdentifiedChannel,
	executorAddress string,
	challengerAddress string,
	outputAddress string,
	submitterAddress string,
	submitTarget ophosttypes.BatchInfo_ChainType,
	submissionInterval time.Duration,
	finalizationPeriod time.Duration,
	submissionStartHeight uint64,
	enableOracle bool,
) (*ophosttypes.MsgCreateBridge, error) {
	// generate ophosthooktypes.PermsMetadata
	// assume that all channels in IBC keeper need to be permitted on OPChild
	// [transfer, nft-transfer, ...]
	permChannels := make([]ophosthooktypes.PortChannelID, 0)
	for _, channel := range identifiedChannels {
		permChannels = append(permChannels, ophosthooktypes.PortChannelID{
			PortID:    channel.Counterparty.PortId,
			ChannelID: channel.Counterparty.ChannelId,
		})
	}

	permsMetadata := ophosthooktypes.PermsMetadata{PermChannels: permChannels}
	permsMetadataJSON, err := json.Marshal(permsMetadata)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to marshal perms metadata")
	}

	// create OpBridgeMessage
	return ophosttypes.NewMsgCreateBridge(
		executorAddress,
		ophosttypes.BridgeConfig{
			Challengers: []string{challengerAddress},
			Proposer:    outputAddress,
			BatchInfo: ophosttypes.BatchInfo{
				Submitter: submitterAddress,
				ChainType: submitTarget,
			},
			SubmissionInterval:    submissionInterval,
			FinalizationPeriod:    finalizationPeriod,
			SubmissionStartHeight: submissionStartHeight,
			Metadata:              permsMetadataJSON,
			OracleEnabled:         enableOracle,
		},
	), nil
}
