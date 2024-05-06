package steps

import (
	"encoding/json"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/initia-labs/OPinit/contrib/launchtools"
	"github.com/initia-labs/OPinit/contrib/launchtools/utils"
	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
	ophosthooktypes "github.com/initia-labs/OPinit/x/ophost/types/hook"
	"github.com/pkg/errors"
)

const (
	BridgeArtifactName = "BRIDGE_ID"
)

// InitializeOpBridge creates OP bridge between OPChild and OPHost
func InitializeOpBridge(
	input launchtools.Input,
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
			input.SystemKeys.Executor.Address,
			input.SystemKeys.Challenger.Address,
			input.SystemKeys.Output.Address,
			input.SystemKeys.Submitter.Address,
			input.OpBridge.SubmitTarget,
			input.OpBridge.SubmissionInterval,
			input.OpBridge.FinalizationPeriod,
			input.OpBridge.SubmissionStartTime,
		)

		ctx.Logger().Info("creating op bridge...", "message", createOpBridgeMessage.String())

		if err != nil {
			return errors.Wrap(err, "failed to create OpBridgeMessage")
		}

		ctx.Logger().Info("broadcasting tx to L1...",
			"from-address", input.SystemKeys.Executor.Address,
		)

		// send createOpBridgeMessage to host (L1)
		res, err := ctx.GetRPCHelperL1().BroadcastTxAndWait(
			input.SystemKeys.Executor.Address,
			input.SystemKeys.Executor.Mnemonic,
			200000,
			sdk.NewCoins(sdk.NewInt64Coin(input.L1Config.Denom, 500000)),
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
	submitTarget string,
	submissionInterval string,
	finalizationPeriod string,
	submissionStartTime time.Time,
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

	interval, err := time.ParseDuration(submissionInterval)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse submission interval %s", submissionInterval)
	}

	period, err := time.ParseDuration(finalizationPeriod)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse finalization period %s", finalizationPeriod)
	}

	// create OpBridgeMessage
	return ophosttypes.NewMsgCreateBridge(
		executorAddress,
		ophosttypes.BridgeConfig{
			Challenger: challengerAddress,
			Proposer:   outputAddress,
			BatchInfo: ophosttypes.BatchInfo{
				Submitter: submitterAddress,
				Chain:     submitTarget,
			},
			SubmissionInterval:  interval,
			FinalizationPeriod:  period,
			SubmissionStartTime: submissionStartTime,
			Metadata:            permsMetadataJSON,
		},
	), nil
}
