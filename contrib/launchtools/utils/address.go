package utils

import (
	"github.com/pkg/errors"

	cmtcrypto "github.com/cometbft/cometbft/crypto"

	"cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"

	ophosttypes "github.com/initia-labs/OPinit/x/ophost/types"
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

func DeriveL1Address(mnemonic string) (string, error) {
	return deriveAddress(mnemonic, L1AddressCodec())
}

func DeriveL2Address(mnemonic string) (string, error) {
	return deriveAddress(mnemonic, L2AddressCodec())
}

func DeriveDAAddress(mnemonic string, chainType ophosttypes.BatchInfo_ChainType) (string, error) {
	var codec address.Codec
	switch chainType {
	case ophosttypes.BatchInfo_CHAIN_TYPE_INITIA:
		codec = L1AddressCodec()
	case ophosttypes.BatchInfo_CHAIN_TYPE_CELESTIA:
		codec = CelestiaAddressCodec()
	default:
		return "", errors.New("unsupported chain type")
	}
	return deriveAddress(mnemonic, codec)
}

func deriveAddress(mnemonic string, codec address.Codec) (string, error) {
	addrBz, err := deriveAddressBz(mnemonic)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert address to bech32")
	}
	return codec.BytesToString(addrBz)
}

func deriveAddressBz(mnemonic string) (cmtcrypto.Address, error) {
	algo := hd.Secp256k1
	derivedPriv, err := algo.Derive()(
		mnemonic,
		keyring.DefaultBIP39Passphrase,
		sdk.GetConfig().GetFullBIP44Path(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to derive private key")
	}

	privKey := algo.Generate()(derivedPriv)
	return privKey.PubKey().Address(), nil
}
