import { KoaController } from 'koa-joi-controllers'
import BatchController from './BatchController'
import { OutputController } from './OutputController'
import { TxController } from './TxController'


export const executorController = [
  OutputController,
  TxController
].map((prototype) =>  new prototype()) as KoaController[]

export const batchController = [
  BatchController,
].map((prototype) =>  new prototype()) as KoaController[]


