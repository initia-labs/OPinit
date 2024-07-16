package opchild

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	opchildv1 "github.com/initia-labs/OPinit/api/opinit/opchild/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: opchildv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Validator",
					Use:       "validator [validator-addr]",
					Short:     "Query a validator",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "validator_addr"},
					},
				},
				{
					RpcMethod: "Validators",
					Use:       "validators",
					Short:     "Query for all validators",
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current opchild parameters information",
				},
				{
					RpcMethod: "NextL1Sequence",
					Use:       "next-l1-sequence",
					Short:     "Query the next l1 sequence",
				},
				{
					RpcMethod: "NextL2Sequence",
					Use:       "next-l2-sequence",
					Short:     "Query the next l2 sequence",
				},
			},
			EnhanceCustomCommand: true, // We still have manual commands in gov that we want to keep
		},
	}
}
