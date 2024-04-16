import { ExecutorDepositTxEntity } from '../../orm'
import { getDB } from '../../lib/db'

export interface GetDepositTxListParam {
  sequence?: number;
  address?: string;
  offset?: number;
  limit: number;
  descending: string;
}

export interface GetDepositTxListResponse {
  count?: number;
  next?: number;
  limit: number;
  depositTxList: ExecutorDepositTxEntity[];
}

export async function getDepositTxList(
  param: GetDepositTxListParam
): Promise<GetDepositTxListResponse> {
  const [db] = getDB()
  const queryRunner = db.createQueryRunner('slave')
  try {
    const offset = param.offset ?? 0
    const order = param.descending == 'true' ? 'DESC' : 'ASC'

    const qb = queryRunner.manager.createQueryBuilder(
      ExecutorDepositTxEntity,
      'tx'
    )

    if (param.sequence) {
      qb.andWhere('tx.sequence = :sequence', { sequence: param.sequence })
    }

    if (param.address) {
      qb.andWhere('tx.sender = :sender', { sender: param.address })
    }

    const depositTxList = await qb
      .orderBy('tx.sequence', order)
      .skip(offset * param.limit)
      .take(param.limit)
      .getMany()

    const count = await qb.getCount()
    let next: number | undefined

    if (count > (offset + 1) * param.limit) {
      next = offset + 1
    }

    return {
      count,
      next,
      limit: param.limit,
      depositTxList
    }
  } finally {
    queryRunner.release()
  }
}
