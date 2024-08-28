package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/store/rootmulti"
	v1 "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
	"github.com/cometbft/cometbft/libs/bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// the state key for withdrawal commitment
func WithdrawalCommitmentKey(sequence uint64) []byte {
	prefixLen := len(WithdrawalCommitmentPrefix)
	buf := make([]byte, prefixLen+8)
	copy(buf, WithdrawalCommitmentPrefix)
	_, err := collections.Uint64Key.Encode(buf[prefixLen:], sequence)
	if err != nil {
		panic(err)
	}

	return buf
}

// CommitWithdrawal returns the withdrwwal commitment bytes. The commitment consists of:
// sha256_hash(receiver_address || l1_amount)
// from a given packet. This results in a fixed length preimage.
func CommitWithdrawal(sequence uint64, receiver string, amount sdk.Coin) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, sequence)
	buf = append(buf, []byte(receiver)...)
	buf = append(buf, amount.String()...)
	hash := sha256.Sum256(buf)
	return hash[:]
}

func VerifyCommitment(appHash []byte, l2SequenceNumber uint64, receiver string, amount sdk.Coin, commitmentProof *v1.ProofOps) error {
	key := WithdrawalCommitmentKey(l2SequenceNumber)
	keyPath := fmt.Sprintf("/%s/x:%s", StoreKey, hex.EncodeToString(key))
	commitment := CommitWithdrawal(l2SequenceNumber, receiver, amount)
	proofOps := NewProofOpsFromProto(commitmentProof)
	return rootmulti.DefaultProofRuntime().VerifyValue(proofOps, appHash, keyPath, commitment)
}

func VerifyAppHash(blockHash, appHash []byte, proof *v1.Proof) error {
	mp, err := NewProofFromProto(proof)
	if err != nil {
		return err
	}

	return mp.Verify(blockHash, cdcEncode(bytes.HexBytes(appHash)))
}
