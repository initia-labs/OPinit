package ophost

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	ophostv1 "github.com/initia-labs/OPinit/api/opinit/ophost/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: ophostv1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Bridge",
					Use:       "bridge [bridge-id]",
					Short:     "Get bridge info",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "bridge_id"},
					},
				},
				{
					RpcMethod: "Bridges",
					Use:       "bridges",
					Short:     "Get bridges info",
				},
				{
					RpcMethod: "Claimed",
					Use:       "claimed [bridge-id] [withdrawal-hash]",
					Short:     "Query whether a withdrawal has been claimed",
				},
				{
					RpcMethod: "NextL1Sequence",
					Use:       "next_l1_sequence [bridge-id]",
					Short:     "Get the next l1 sequence",
				},
				{
					RpcMethod: "TokenPairByL1Denom",
					Use:       "token_pair_by_l1_denom [bridge-id]",
					Short:     "Get the token pair by l1 denom",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "bridge_id"},
					},
				},
				{
					RpcMethod: "TokenPairByL2Denom",
					Use:       "token_pair_by_l2_denom [bridge-id]",
					Short:     "Get the token pair by l2 denom",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "bridge_id"},
					},
				},
				{
					RpcMethod: "TokenPairs",
					Use:       "token_pairs [bridge-id]",
					Short:     "Get the token pairs",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "bridge_id"},
					},
				},
				{
					RpcMethod: "LastFinalizedOutput",
					Use:       "last_finalized_output [bridge-id]",
					Short:     "Get last finalized output proposal",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "bridge_id"},
					},
				},
				{
					RpcMethod: "OutputProposal",
					Use:       "output_proposal [bridge-id] [output_index]",
					Short:     "Get output proposal",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "bridge_id"},
						{ProtoField: "output_index"},
					},
				},
				{
					RpcMethod: "OutputProposals",
					Use:       "output_proposals [bridge-id]",
					Short:     "Get output proposals",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "bridge_id"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Returns the ophost module's parameters",
				},
			},
			EnhanceCustomCommand: true, // We still have manual commands in gov that we want to keep
		},
	}
}
