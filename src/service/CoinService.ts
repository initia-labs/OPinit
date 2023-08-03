import { CoinEntity } from 'orm';
import { getDB } from 'worker/bridgeExecutor/db';
import { APIError, ErrorTypes } from 'lib/error';

export async function getCoin(coinType: string): Promise<CoinEntity> {
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

    return coin;
  } finally {
    queryRunner.release();
  }
}
