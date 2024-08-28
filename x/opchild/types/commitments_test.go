package types

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/tmhash"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	comettypes "github.com/cometbft/cometbft/types"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store/iavl"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_VerifyCommitment(t *testing.T) {
	db := dbm.NewMemDB()
	store := rootmulti.NewStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	iavlStoreKey := storetypes.NewKVStoreKey(StoreKey)

	store.MountStoreWithDB(iavlStoreKey, storetypes.StoreTypeIAVL, nil)
	require.NoError(t, store.LoadVersion(0))

	sequence := uint64(10)
	commitmentKey := WithdrawalCommitmentKey(sequence)

	recipient := "recipient"
	amount := sdk.NewInt64Coin("uinit", 100)
	commitment := CommitWithdrawal(sequence, recipient, amount)

	iavlStore := store.GetCommitStore(iavlStoreKey).(*iavl.Store)
	iavlStore.Set(commitmentKey, commitment)
	cid := store.Commit()

	// Get Proof
	// same with curl https://rpc.initia.xyz/abci_query\?path\="\"store/opchild/key\""\&data=0xcommitmentkey\&prove=true
	res, err := store.Query(&storetypes.RequestQuery{
		Path:  fmt.Sprintf("/%s/key", StoreKey), // required path to get key/value+proof
		Data:  commitmentKey,
		Prove: true,
	})
	require.NoError(t, err)
	require.NotNil(t, res.ProofOps)

	// Verify proof.
	err = VerifyCommitment(cid.Hash, sequence, recipient, amount, NewProtoFromProofOps(res.ProofOps))
	require.Nil(t, err)
}

func Test_VerifyAppHash(t *testing.T) {
	block := makeRandBlock(t)
	header := block.Header
	appHashProof := NewAppHashProof(&header)
	require.NotNil(t, appHashProof)

	err := VerifyAppHash(block.Hash(), header.AppHash, appHashProof)
	require.NoError(t, err)
}

func makeRandBlock(t *testing.T) *comettypes.Block {
	txs := []comettypes.Tx{comettypes.Tx("foo"), comettypes.Tx("bar")}
	lastID := makeBlockIDRandom()
	h := int64(3)
	voteSet, valSet, vals := randVoteSet(h-1, 1, cmtproto.PrecommitType, 10, 1, false)
	extCommit, err := comettypes.MakeExtCommit(lastID, h-1, 1, voteSet, vals, time.Now(), false)
	require.NoError(t, err)

	ev, err := comettypes.NewMockDuplicateVoteEvidenceWithValidator(h, time.Now(), vals[0], "block-test-chain")
	require.NoError(t, err)
	evList := []comettypes.Evidence{ev}

	block := comettypes.MakeBlock(h, txs, extCommit.ToCommit(), evList)
	block.ValidatorsHash = valSet.Hash()
	block.AppHash = tmhash.Sum([]byte("app_hash"))

	return block
}

func makeBlockIDRandom() comettypes.BlockID {
	var (
		blockHash   = make([]byte, tmhash.Size)
		partSetHash = make([]byte, tmhash.Size)
	)
	rand.Read(blockHash)   //nolint: errcheck // ignore errcheck for read
	rand.Read(partSetHash) //nolint: errcheck // ignore errcheck for read
	return comettypes.BlockID{Hash: blockHash, PartSetHeader: comettypes.PartSetHeader{Total: 123, Hash: partSetHash}}
}

// NOTE: privValidators are in order
func randVoteSet(
	height int64,
	round int32,
	signedMsgType cmtproto.SignedMsgType,
	numValidators int,
	votingPower int64,
	extEnabled bool,
) (*comettypes.VoteSet, *comettypes.ValidatorSet, []comettypes.PrivValidator) {
	valSet, privValidators := comettypes.RandValidatorSet(numValidators, votingPower)
	if extEnabled {
		if signedMsgType != cmtproto.PrecommitType {
			return nil, nil, nil
		}
		return comettypes.NewExtendedVoteSet("test_chain_id", height, round, signedMsgType, valSet), valSet, privValidators
	}
	return comettypes.NewVoteSet("test_chain_id", height, round, signedMsgType, valSet), valSet, privValidators
}
