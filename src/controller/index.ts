import { KoaController } from 'koa-joi-controllers';
import BatchController from './batch/BatchController';
import { OutputController } from './executor/OutputController';
import { TxController } from './executor/TxController';

export const executorController = [OutputController, TxController].map(
  (prototype) => new prototype()
) as KoaController[];

export const batchController = [BatchController].map(
  (prototype) => new prototype()
) as KoaController[];
