package opchild

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	opchildv1 "github.com/initia-labs/OPinit/v1/api/opinit/opchild/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              opchildv1.Query_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true,
		},
	}
}
