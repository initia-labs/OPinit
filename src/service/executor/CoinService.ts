import { ExecutorCoinEntity } from 'orm';
import { getDB } from 'worker/bridgeExecutor/db';
import { APIError, ErrorTypes } from 'lib/error';

export interface GetCoinResponse {
  coin: ExecutorCoinEntity;
}

export interface GetAllCoinsResponse {
  coins: ExecutorCoinEntity[];
}

export async function getCoin(metadata: string): Promise<GetCoinResponse> {
  const [db] = getDB();
  const queryRunner = db.createQueryRunner('slave');
  try {
    const qb = queryRunner.manager
      .createQueryBuilder(ExecutorCoinEntity, 'coin')
      .where('coin.l1Metadata = :metadata', { metadata });

    const coin = await qb.getOne();

    if (!coin) {
      throw new APIError(ErrorTypes.NOT_FOUND_ERROR);
    }

    return {
      coin: coin
    };
  } finally {
    queryRunner.release();
  }
}

export async function getAllCoins(): Promise<GetAllCoinsResponse> {
  const [db] = getDB();
  const queryRunner = db.createQueryRunner('slave');
  try {
    const qb = queryRunner.manager.createQueryBuilder(
      ExecutorCoinEntity,
      'coin'
    );

    const coins = await qb.getMany();

    if (!coins) {
      throw new APIError(ErrorTypes.NOT_FOUND_ERROR);
    }

    return {
      coins: coins
    };
  } finally {
    queryRunner.release();
  }
}
