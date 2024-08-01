package keeper_test

import (
	"testing"
	"time"

	"github.com/initia-labs/OPinit/x/ophost/types"
	"github.com/stretchr/testify/require"
)

func Test_SetGetBatchInfo(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	batchInfo1 := types.BatchInfo{
		Submitter: addrsStr[0],
		ChainType: types.BatchInfo_CHAIN_TYPE_INITIA,
	}
	output1 := types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 100,
	}

	err := input.OPHostKeeper.SetBatchInfo(ctx, 1, batchInfo1, output1)
	require.NoError(t, err)

	batchInfo2 := types.BatchInfo{
		Submitter: addrsStr[1],
		ChainType: types.BatchInfo_CHAIN_TYPE_INITIA,
	}
	output2 := types.Output{
		OutputRoot:    []byte{4, 5, 6},
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 200,
	}

	err = input.OPHostKeeper.SetBatchInfo(ctx, 1, batchInfo2, output2)
	require.NoError(t, err)

	batchInfo3 := types.BatchInfo{
		Submitter: addrsStr[0],
		ChainType: types.BatchInfo_CHAIN_TYPE_CELESTIA,
	}
	output3 := types.Output{
		OutputRoot:    []byte{1, 2, 3},
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 300,
	}

	err = input.OPHostKeeper.SetBatchInfo(ctx, 2, batchInfo3, output3)
	require.NoError(t, err)

	batchInfo4 := types.BatchInfo{
		Submitter: addrsStr[1],
		ChainType: types.BatchInfo_CHAIN_TYPE_CELESTIA,
	}
	output4 := types.Output{
		OutputRoot:    []byte{4, 5, 6},
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 400,
	}

	err = input.OPHostKeeper.SetBatchInfo(ctx, 2, batchInfo4, output4)
	require.NoError(t, err)

	batchInfo5 := types.BatchInfo{
		Submitter: addrsStr[1],
		ChainType: types.BatchInfo_CHAIN_TYPE_CELESTIA,
	}
	output5 := types.Output{
		OutputRoot:    []byte{4, 5, 6},
		L1BlockTime:   time.Now().UTC(),
		L2BlockNumber: 500,
	}

	err = input.OPHostKeeper.SetBatchInfo(ctx, 2, batchInfo5, output5)
	require.NoError(t, err)

	batchInfos, err := input.OPHostKeeper.GetAllBatchInfos(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, batchInfo1, batchInfos[0].BatchInfo)
	require.Equal(t, output1, batchInfos[0].Output)
	require.Equal(t, batchInfo2, batchInfos[1].BatchInfo)
	require.Equal(t, output2, batchInfos[1].Output)

	batchInfos, err = input.OPHostKeeper.GetAllBatchInfos(ctx, 2)
	require.NoError(t, err)
	require.Equal(t, batchInfo3, batchInfos[0].BatchInfo)
	require.Equal(t, output3, batchInfos[0].Output)
	require.Equal(t, batchInfo4, batchInfos[1].BatchInfo)
	require.Equal(t, output4, batchInfos[1].Output)
	require.Equal(t, batchInfo5, batchInfos[2].BatchInfo)
	require.Equal(t, output5, batchInfos[2].Output)
}
