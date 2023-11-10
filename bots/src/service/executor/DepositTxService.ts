import { ExecutorDepositTxEntity } from 'orm';
import { getDB } from 'worker/bridgeExecutor/db';
import { APIError, ErrorTypes } from 'lib/error';

export interface GetDepositTxResponse {
  depositTx: ExecutorDepositTxEntity;
}

export async function getDepositTx(
  bridgeId: string,
  sequence: number
): Promise<GetDepositTxResponse> {
  const [db] = getDB();
  const queryRunner = db.createQueryRunner('slave');
  try {
    const qb = queryRunner.manager
      .createQueryBuilder(ExecutorDepositTxEntity, 'tx')
      .where('tx.bridge_id = :bridgeId', { bridgeId })
      .andWhere('tx.sequence = :sequence', { sequence });

    const depositTx = await qb.getOne();

    if (!depositTx) {
      throw new APIError(ErrorTypes.NOT_FOUND_ERROR);
    }

    return {
      depositTx
    };
  } finally {
    queryRunner.release();
  }
}
