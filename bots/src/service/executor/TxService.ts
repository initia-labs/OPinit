import { ExecutorWithdrawalTxEntity } from 'orm';
import { getDB } from 'worker/bridgeExecutor/db';
import { APIError, ErrorTypes } from 'lib/error';

export async function getTx(
  metadata: string,
  sequence: number
): Promise<ExecutorWithdrawalTxEntity> {
  const [db] = getDB();
  const queryRunner = db.createQueryRunner('slave');
  try {
    const qb = queryRunner.manager
      .createQueryBuilder(ExecutorWithdrawalTxEntity, 'tx')
      .where('tx.metadata = :metadata', { metadata })
      .andWhere('tx.sequence = :sequence', { sequence });

    const tx = await qb.getOne();

    if (!tx) {
      throw new APIError(ErrorTypes.NOT_FOUND_ERROR);
    }

    return tx;
  } finally {
    queryRunner.release();
  }
}
