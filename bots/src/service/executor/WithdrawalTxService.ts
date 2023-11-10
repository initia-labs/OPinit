import { ExecutorWithdrawalTxEntity } from 'orm';
import { getDB } from 'worker/bridgeExecutor/db';
import { APIError, ErrorTypes } from 'lib/error';

export interface GetWithdrawalTxResponse {
  withdrawalTx: ExecutorWithdrawalTxEntity;
}

export async function getWithdrawalTx(
  bridgeId: string,
  sequence: number
): Promise<GetWithdrawalTxResponse> {
  const [db] = getDB();
  const queryRunner = db.createQueryRunner('slave');
  try {
    const qb = queryRunner.manager
      .createQueryBuilder(ExecutorWithdrawalTxEntity, 'tx')
      .where('tx.bridge_id = :bridgeId', { bridgeId })
      .andWhere('tx.sequence = :sequence', { sequence });

    const withdrawalTx = await qb.getOne();

    if (!withdrawalTx) {
      throw new APIError(ErrorTypes.NOT_FOUND_ERROR);
    }

    return {
      withdrawalTx
    };
  } finally {
    queryRunner.release();
  }
}
