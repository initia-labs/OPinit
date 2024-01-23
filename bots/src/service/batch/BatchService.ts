import { RecordEntity } from 'orm';
import { APIError, ErrorTypes } from 'lib/error';
import { getDB } from 'worker/batchSubmitter/db';
import { decompress } from 'lib/compressor';

interface GetBatchResponse {
  bridgeId: number;
  batchIndex: number;
  batch: string[];
}

export async function getBatch(batchIndex: number): Promise<GetBatchResponse> {
  const [db] = getDB();
  const queryRunner = db.createQueryRunner('slave');

  try {
    const qb = queryRunner.manager
      .createQueryBuilder(RecordEntity, 'record')
      .where('record.batchIndex = :batchIndex', { batchIndex });

    const batch = await qb.getOne();

    if (!batch) {
      throw new APIError(ErrorTypes.NOT_FOUND_ERROR);
    }

    return {
      bridgeId: batch.bridgeId,
      batchIndex: batch.batchIndex,
      batch: decompress(batch.batch)
    };
  } finally {
    queryRunner.release();
  }
}
