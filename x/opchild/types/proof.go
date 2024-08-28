package types

import (
	"bytes"

	v1 "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
	"github.com/cometbft/cometbft/crypto/merkle"
	cmtcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmttypes "github.com/cometbft/cometbft/types"

	gogotypes "github.com/cosmos/gogoproto/types"
)

func NewProofFromProto(p *v1.Proof) (*merkle.Proof, error) {
	proof := &merkle.Proof{
		Total:    p.Total,
		Index:    p.Index,
		LeafHash: p.LeafHash,
		Aunts:    p.Aunts,
	}

	return proof, proof.ValidateBasic()
}

func NewProofOpsFromProto(p *v1.ProofOps) *cmtcrypto.ProofOps {
	ops := make([]cmtcrypto.ProofOp, len(p.Ops))
	for i, op := range p.Ops {
		ops[i] = cmtcrypto.ProofOp(op)
	}

	return &cmtcrypto.ProofOps{Ops: ops}
}

func NewProtoFromProof(p *merkle.Proof) *v1.Proof {
	return &v1.Proof{
		Total:    p.Total,
		Index:    p.Index,
		LeafHash: p.LeafHash,
		Aunts:    p.Aunts,
	}
}

func NewProtoFromProofOps(p *cmtcrypto.ProofOps) *v1.ProofOps {
	ops := make([]v1.ProofOp, len(p.Ops))
	for i, op := range p.Ops {
		ops[i] = v1.ProofOp(op)
	}

	return &v1.ProofOps{Ops: ops}
}

func NewAppHashProof(h *cmttypes.Header) *v1.Proof {
	if h == nil || len(h.ValidatorsHash) == 0 {
		return nil
	}
	hbz, err := h.Version.Marshal()
	if err != nil {
		return nil
	}

	pbt, err := gogotypes.StdTimeMarshal(h.Time)
	if err != nil {
		return nil
	}

	pbbi := h.LastBlockID.ToProto()
	bzbi, err := pbbi.Marshal()
	if err != nil {
		return nil
	}

	rootHash, proofs := merkle.ProofsFromByteSlices([][]byte{
		hbz,
		cdcEncode(h.ChainID),
		cdcEncode(h.Height),
		pbt,
		bzbi,
		cdcEncode(h.LastCommitHash),
		cdcEncode(h.DataHash),
		cdcEncode(h.ValidatorsHash),
		cdcEncode(h.NextValidatorsHash),
		cdcEncode(h.ConsensusHash),
		cdcEncode(h.AppHash),
		cdcEncode(h.LastResultsHash),
		cdcEncode(h.EvidenceHash),
		cdcEncode(h.ProposerAddress),
	})
	if !bytes.Equal(rootHash, h.Hash().Bytes()) {
		return nil
	}

	return NewProtoFromProof(proofs[10])
}
