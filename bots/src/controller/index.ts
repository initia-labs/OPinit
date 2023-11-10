import { KoaController } from 'koa-joi-controllers';
import BatchController from './batch/BatchController';
import { OutputController } from './executor/OutputController';
import { WithdrawalTxController } from './executor/WithdrawalTxController';
import { DepositTxController } from './executor/DepositTxController';

export const executorController = [
  OutputController,
  WithdrawalTxController,
  DepositTxController
].map((prototype) => new prototype()) as KoaController[];

export const batchController = [BatchController].map(
  (prototype) => new prototype()
) as KoaController[];
