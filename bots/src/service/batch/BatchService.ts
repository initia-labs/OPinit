import { RecordEntity } from '../../orm'
import { APIError, ErrorTypes } from '../../lib/error'
import { getDB } from '../../lib/db'
import { config } from '../../config'

interface L1BatchInfo {
  type: 'l1';
  dataPaths: {
    index: number;
    txHash: string;
  }[];
}

interface CelestiaBatchInfo {
  type: 'celestia';
  dataPaths: {
    index: number;
    height: number;
    commitment: string;
  }[];
}

type BatchInfo = L1BatchInfo | CelestiaBatchInfo

interface GetBatchResponse {
  bridgeId: number;
  batchIndex: number;
  batchInfo: BatchInfo;
}

export async function getBatch(batchIndex: number): Promise<GetBatchResponse> {
  const [db] = getDB()
  const queryRunner = db.createQueryRunner('slave')

  try {
    const qb = queryRunner.manager
      .createQueryBuilder(RecordEntity, 'record')
      .where('record.batch_index = :batchIndex', { batchIndex })

    const batch = await qb.getOne()

    if (!batch) {
      throw new APIError(ErrorTypes.NOT_FOUND_ERROR)
    }

    let batchInfo: BatchInfo
    if (config.PUBLISH_BATCH_TARGET === 'l1') {
      batchInfo = {
        type: 'l1',
        dataPaths: batch.batchInfo.map((txHash, index) => ({ index, txHash }))
      }
    } else if (config.PUBLISH_BATCH_TARGET === 'celestia') {
      batchInfo = {
        type: 'celestia',
        dataPaths: batch.batchInfo.map((path, index) => {
          const [height, commitment] = path.split('::')
          return { index, height: Number(height), commitment }
        })
      }
    } else {
      throw new APIError(ErrorTypes.API_ERROR)
    }

    return {
      bridgeId: batch.bridgeId,
      batchIndex: batch.batchIndex,
      batchInfo
    }
  } finally {
    queryRunner.release()
  }
}
