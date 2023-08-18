import { CoinEntity } from 'orm';
import { getDB } from 'worker/bridgeExecutor/db';
import { APIError, ErrorTypes } from 'lib/error';

export interface GetCoinResponse {
  coin: CoinEntity;
}

export interface GetAllCoinsResponse {
  coins: CoinEntity[];
}

export async function getCoin(coinType: string): Promise<GetCoinResponse> {
  const [db] = getDB();
  const queryRunner = db.createQueryRunner('slave');
  try {
    const qb = queryRunner.manager
      .createQueryBuilder(CoinEntity, 'coin')
      .where('coin.l1StructTag = :coinType', { coinType });

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
    const qb = queryRunner.manager.createQueryBuilder(CoinEntity, 'coin');

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
