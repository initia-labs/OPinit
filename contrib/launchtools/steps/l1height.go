package steps

import (
	"fmt"

	"github.com/initia-labs/OPinit/v1/contrib/launchtools"
	"github.com/pkg/errors"
)

var _ launchtools.LauncherStepFuncFactory[*launchtools.Config] = GetL1Height

// GetL1Height gets the height of the L1 chain. Useful to determine opinit's initial monitor height.
func GetL1Height(_ *launchtools.Config) launchtools.LauncherStepFunc {
	const OutputName = "EXECUTOR_L1_MONITOR_HEIGHT"

	return func(ctx launchtools.Launcher) error {
		if ctx.GetRPCHelperL1() == nil {
			return errors.New("RPC helper for L1 not initialized")
		}

		status, err := ctx.GetRPCHelperL1().GetStatus()
		if err != nil {
			return errors.Wrapf(err, "failed to get status from L1")
		}

		ctx.Logger().Info("L1 chain status", "height", status.SyncInfo.LatestBlockHeight)

		return ctx.WriteOutput(OutputName, fmt.Sprintf("%d", status.SyncInfo.LatestBlockHeight))
	}
}
