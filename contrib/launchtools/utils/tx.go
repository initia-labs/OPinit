package utils

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/pkg/errors"
)

var (
	algo = hd.Secp256k1
)

// SignTxOffline signs a transaction offline using the provided mnemonic.
// This is useful for signing transactions in an offline environment.
func SignTxOffline(
	ctx *client.Context,
	mnemonic string,
	gasLimit uint64,
	accountNumber uint64,
	sequence uint64,
	fee sdk.Coins,
	msgs ...sdk.Msg,
) (xauthsigning.Tx, error) {
	// Cosmos requires sigv2 to be provided in 2 steps
	txbldr := ctx.TxConfig.NewTxBuilder()
	txbldr.SetFeeAmount(fee)
	txbldr.SetGasLimit(gasLimit)
	if err := txbldr.SetMsgs(msgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to set messages")
	}

	// derive private key using mnemonic
	// Note: assumes default passphrase + default bip44 path set in the config
	derivedPriv, err := algo.Derive()(mnemonic, keyring.DefaultBIP39Passphrase, sdk.GetConfig().GetFullBIP44Path())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to derive private key")
	}
	privKey := algo.Generate()(derivedPriv)

	// create "empty" signature and assign first
	if err := txbldr.SetSignatures(signing.SignatureV2{
		Sequence: sequence,
		PubKey:   privKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
			Signature: nil,
		},
	}); err != nil {
		return nil, errors.Wrapf(err, "failed to set signature")
	}

	// using the assigned auth info, sign the transaction (actually this time
	sigV2, err := tx.SignWithPrivKey(
		context.Background(), // not required
		signing.SignMode_SIGN_MODE_DIRECT,
		xauthsigning.SignerData{
			ChainID:       ctx.ChainID,
			AccountNumber: accountNumber,
			Sequence:      sequence,
		},
		txbldr,
		privKey,
		ctx.TxConfig,
		sequence,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to sign with private key")
	}

	// set signature
	if err := txbldr.SetSignatures(sigV2); err != nil {
		return nil, errors.Wrapf(err, "failed to set signature")
	}

	// validate if the transaction is valid
	// if err := txbldr.GetTx().ValidateBasic(); err != nil {
	// 	return nil, errors.Wrapf(err, "failed to validate basic")
	// }

	// return if no problem
	return txbldr.GetTx(), nil
}
