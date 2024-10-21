package utils

import (
	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
)

func L1AddressCodec() address.Codec {
	return authcodec.NewBech32Codec("init")
}

func L2AddressCodec() address.Codec {
	return authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
}

func CelestiaAddressCodec() address.Codec {
	return authcodec.NewBech32Codec("celestia")
}

func HackBech32Prefix(prefix string) func() {
	originPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	originPubPrefix := sdk.GetConfig().GetBech32AccountPubPrefix()
	sdk.GetConfig().SetBech32PrefixForAccount(prefix, prefix+"pub")

	return func() {
		sdk.GetConfig().SetBech32PrefixForAccount(originPrefix, originPubPrefix)
	}
}
