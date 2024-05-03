package steps

import (
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	tmclient "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/initia-labs/OPinit/contrib/launchtools"
	"github.com/pkg/errors"
)

// EnableOracle enables the OP oracle (?)
func EnableOracle(input launchtools.Input) launchtools.LauncherStepFunc {
	return func(ctx launchtools.Launcher) error {
		ctx.Logger().Info("enabling oracle")

		// scan all states from IBC - get all established states
		res, err := ctx.App().GetIBCKeeper().ClientStates(
			ctx.QueryContext(),
			&clienttypes.QueryClientStatesRequest{},
		)
		if err != nil {
			return errors.Wrapf(err, "failed to query client states")
		}

		var l1ClientID string

		// among the states, find the one with the chain ID of L1
		for _, st := range res.ClientStates {
			clientState, err := clienttypes.UnpackClientState(st.ClientState)
			if err != nil {
				return errors.Wrapf(err, "failed to unpack client state")
			}
			tcli, ok := clientState.(*tmclient.ClientState)
			if !ok {
				ctx.Logger().Debug("skipping non-tendermint client", "client-type", clientState.String())
				continue
			}

			if tcli.ChainId == input.L1Config.ChainID {
				ctx.Logger().Info("found L1 tendermint client", "client-id", st.ClientId)
				l1ClientID = st.ClientId
				break
			}
		}

		// if  L1 tendermint client is never found, return an error
		if l1ClientID == "" {
			return errors.New("failed to find L1 tendermint client")
		}

		// otherwise write to a file and return
		return ctx.WriteToFile("oracle-client-id", l1ClientID)
	}
}
