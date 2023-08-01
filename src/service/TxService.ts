import { TxEntity } from 'orm'
import { getDB } from 'worker/bridgeExecutor/db'
import { APIError, ErrorTypes } from "lib/error";


export async function getTx(
  coin_type: string,
  sequence: number
): Promise<TxEntity> {
  const [db] = getDB()
  const queryRunner = db.createQueryRunner('slave')
  try {
    const qb = queryRunner.manager
      .createQueryBuilder(TxEntity, 'tx')
      .where('tx.coin_type = :coin_type', { coin_type })
      .andWhere('tx.sequence = :sequence', { sequence })

    const tx = await qb.getOne()

    if (!tx) {
      throw new APIError(ErrorTypes.NOT_FOUND_ERROR)
    }

    return tx
  } finally {
    queryRunner.release()
  }
}