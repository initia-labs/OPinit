package types

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	v1 "github.com/cometbft/cometbft/api/cometbft/crypto/v1"

	"github.com/cosmos/cosmos-sdk/client"
)

type BaseAppQuerier func(_ context.Context, req *abci.RequestQuery) (resp *abci.ResponseQuery, err error)

func QueryCommitmentProof(baseAppQuerier BaseAppQuerier, height int64, commitmentKey []byte) (*v1.ProofOps, error) {
	res, err := baseAppQuerier(context.Background(), &abci.RequestQuery{
		Path:   fmt.Sprintf("/store/%s/key", StoreKey),
		Data:   commitmentKey,
		Height: height,
		Prove:  true,
	})
	if err != nil {
		return nil, err
	}

	return NewProtoFromProofOps(res.ProofOps), nil
}

func QueryAppHashWithProof(clientCtx *client.Context, height int64) ([]byte, *v1.Proof, error) {
	if clientCtx == nil {
		return nil, nil, fmt.Errorf("clientCtx cannot be nil")
	}

	node, err := clientCtx.GetNode()
	if err != nil {
		return nil, nil, err
	}

	block, err := node.Block(context.Background(), &height)
	if err != nil {
		return nil, nil, err
	}

	appHashProof := NewAppHashProof(&block.Block.Header)
	return block.Block.Header.AppHash, appHashProof, nil
}
