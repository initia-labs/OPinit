package ophost

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	ophostv1 "github.com/initia-labs/OPinit/api/opinit/ophost/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              ophostv1.Query_ServiceDesc.ServiceName,
			EnhanceCustomCommand: true, // We still have manual commands in gov that we want to keep
		},
	}
}
