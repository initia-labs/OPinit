import { KoaController } from 'koa-joi-controllers';
import BatchController from './batch/BatchController';
import { OutputController } from './executor/OutputController';
import { TxController } from './executor/TxController';
import CoinController from './executor/CoinController';

export const executorController = [OutputController, TxController, CoinController].map(
  (prototype) => new prototype()
) as KoaController[];

export const batchController = [BatchController].map(
  (prototype) => new prototype()
) as KoaController[];
