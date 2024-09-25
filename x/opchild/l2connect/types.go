package l2connect

import (
	"context"

	"cosmossdk.io/math"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ReservedCPTimestamp = "TIMESTAMP/NANOSECOND"

// ValidatorStore defines the interface contract require for verifying vote
// extension signatures. Typically, this will be implemented by the x/staking
// module, which has knowledge of the CometBFT public key.
type ValidatorStore interface {
	TotalBondedTokens(ctx context.Context) (math.Int, error)
	GetPubKeyByConsAddr(context.Context, sdk.ConsAddress) (cmtprotocrypto.PublicKey, error)
	GetPowerByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (math.Int, error)
}
