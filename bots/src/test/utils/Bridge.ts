import {
  MsgCreateBridge,
  BridgeConfig,
  Duration,
  BatchInfo
} from '@initia/initia.js'
import {
  getDB as getExecutorDB,
  initORM as initExecutorORM
} from '../../worker/bridgeExecutor/db'
import {
  getDB as getChallengerDB,
  initORM as initChallengerORM
} from '../../worker/challenger/db'
import { getDB as getBatchDB, initORM as initBatchORM } from '../../lib/db'
import { DataSource, EntityManager } from 'typeorm'
import {
  ExecutorOutputEntity,
  StateEntity,
  ExecutorWithdrawalTxEntity,
  ExecutorDepositTxEntity,
  ExecutorUnconfirmedTxEntity,
  ChallengerDepositTxEntity,
  ChallengerFinalizeDepositTxEntity,
  ChallengerFinalizeWithdrawalTxEntity,
  ChallengerOutputEntity,
  ChallengerWithdrawalTxEntity,
  ChallengedOutputEntity,
  RecordEntity,
  ChallengeEntity
} from '../../orm'
import { executor, challenger, outputSubmitter } from './helper'
import { sendTx } from '../../lib/tx'

class Bridge {
  executorDB: DataSource
  challengerDB: DataSource
  batchDB: DataSource
  l1BlockHeight: number
  l2BlockHeight: number

  constructor(
    public submissionInterval: number,
    public finalizedTime: number
  ) {}

  async clearDB() {
    // remove and initialize
    await initExecutorORM()
    await initChallengerORM()
    await initBatchORM();

    [this.executorDB] = getExecutorDB();
    [this.challengerDB] = getChallengerDB();
    [this.batchDB] = getBatchDB()

    await this.executorDB.transaction(async (manager: EntityManager) => {
      await manager.getRepository(StateEntity).clear()
      await manager.getRepository(ExecutorWithdrawalTxEntity).clear()
      await manager.getRepository(ExecutorOutputEntity).clear()
      await manager.getRepository(ExecutorDepositTxEntity).clear()
      await manager.getRepository(ExecutorUnconfirmedTxEntity).clear()
    })

    await this.challengerDB.transaction(async (manager: EntityManager) => {
      await manager.getRepository(ChallengerDepositTxEntity).clear()
      await manager.getRepository(ChallengerFinalizeDepositTxEntity).clear()
      await manager.getRepository(ChallengerFinalizeWithdrawalTxEntity).clear()
      await manager.getRepository(ChallengerOutputEntity).clear()
      await manager.getRepository(ChallengerWithdrawalTxEntity).clear()
      await manager.getRepository(ChallengedOutputEntity).clear()
      await manager.getRepository(ChallengeEntity).clear()
    })

    await this.batchDB.transaction(async (manager: EntityManager) => {
      await manager.getRepository(RecordEntity).clear()
    })
  }

  MsgCreateBridge(
    submissionInterval: number,
    finalizedTime: number,
    metadata: string
  ) {
    const bridgeConfig = new BridgeConfig(
      challenger.key.accAddress,
      outputSubmitter.key.accAddress,
      new BatchInfo('submitter', 'chain'),
      Duration.fromString(submissionInterval.toString()),
      Duration.fromString(finalizedTime.toString()),
      new Date(),
      metadata
    )
    return new MsgCreateBridge(executor.key.accAddress, bridgeConfig)
  }

  async tx(metadata: string) {
    const msgs = [
      this.MsgCreateBridge(
        this.submissionInterval,
        this.finalizedTime,
        metadata
      )
    ]

    return await sendTx(executor, msgs)
  }
}

export default Bridge
