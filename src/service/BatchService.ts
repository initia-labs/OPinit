import { RecordEntity } from "orm";
import { APIError, ErrorTypes } from "lib/error";
import { getDB } from "worker/bridgeExecutor/db";
import { decompressor } from "lib/compressor";

interface GetBatchResponse {
    l2Id: string
    batchIndex: number
    batch: string[]
}
  
export async function getBatch(batchIndex: number): Promise<GetBatchResponse>{
    const [db] = getDB()
    const queryRunner = db.createQueryRunner('slave')
  
    try{
      const qb = queryRunner.manager
        .createQueryBuilder(RecordEntity, 'record')
        .where('record.batchIndex = :batchIndex', { batchIndex })
  
      const batch = await qb.getOne()
  
      if (!batch) {
        throw new APIError(ErrorTypes.NOT_FOUND_ERROR)
      }
  
      return {
        l2Id: batch.l2Id,
        batchIndex: batch.batchIndex,
        batch: decompressor(batch.batch)
      }
    }finally {
      queryRunner.release()
    }
}