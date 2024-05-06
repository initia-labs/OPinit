package steps

import (
	"fmt"
	"github.com/initia-labs/OPinit/contrib/launchtools"
	"github.com/pkg/errors"
)

var _ launchtools.LauncherStepFuncFactory[launchtools.Input] = GetL1Height

// GetL2Height gets the height of the L2 chain. Useful to determine opinit's initial monitor height.
func GetL2Height(_ launchtools.Input) launchtools.LauncherStepFunc {
	const OutputName = "EXECUTOR_L2_MONITOR_HEIGHT"

	return func(ctx launchtools.Launcher) error {
		if ctx.GetRPCHelperL1() == nil {
			return errors.New("RPC helper for L1 not initialized")
		}

		status, err := ctx.GetRPCHelperL2().GetStatus()
		if err != nil {
			return errors.Wrapf(err, "failed to get status from L1")
		}

		ctx.Logger().Info("L2 chain status", "height", status.SyncInfo.LatestBlockHeight)

		return ctx.WriteOutput(OutputName, fmt.Sprintf("%d", status.SyncInfo.LatestBlockHeight))
	}
}
