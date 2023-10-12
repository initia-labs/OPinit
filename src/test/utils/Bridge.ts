import { MsgPublish, MsgExecute } from '@initia/initia.js';
import * as fs from 'fs';
import * as path from 'path';
import { getDB, initORM } from 'worker/bridgeExecutor/db';
import { DataSource, EntityManager } from 'typeorm';
import {
  ExecutorCoinEntity,
  ExecutorOutputEntity,
  StateEntity,
  ExecutorWithdrawalTxEntity
} from 'orm';
import { getConfig } from 'config';
import { build, executor, challenger, outputSubmitter, bcs } from './helper';
import { sendTx } from 'lib/tx';

const config = getConfig();

class Bridge {
  db: DataSource;
  submissionInterval: number;
  finalizedTime: number;
  l2StartBlockHeight: number;
  l1BlockHeight: number;
  l2BlockHeight: number;
  l2id: string;
  moduleName: string;
  contractDir: string;

  constructor(
    submissionInterval: number,
    finalizedTime: number,
    l2StartBlockHeight: number,
    l2id: string,
    contractDir: string
  ) {
    [this.db] = getDB();
    this.submissionInterval = submissionInterval;
    this.finalizedTime = finalizedTime;
    this.l2StartBlockHeight = l2StartBlockHeight;
    this.l2id = l2id;
    this.moduleName = this.l2id.split('::')[1];
    this.contractDir = contractDir;
  }

  async init() {
    await this.setDB();
    this.updateL2ID();
  }

  async setDB() {
    const l1Monitor = `executor_l1_monitor`;
    const l2Monitor = `executor_l2_monitor`;
    this.l1BlockHeight = parseInt(
      (await config.l1lcd.tendermint.blockInfo()).block.header.height
    );

    this.l2BlockHeight = parseInt(
      (await config.l2lcd.tendermint.blockInfo()).block.header.height
    );
    this.l2BlockHeight = Math.floor(this.l2BlockHeight / 100) * 100;

    // remove and initialize
    await this.db.transaction(
      async (transactionalEntityManager: EntityManager) => {
        await transactionalEntityManager.getRepository(StateEntity).clear();
        await transactionalEntityManager
          .getRepository(ExecutorWithdrawalTxEntity)
          .clear();
        await transactionalEntityManager
          .getRepository(ExecutorCoinEntity)
          .clear();
        await transactionalEntityManager
          .getRepository(ExecutorOutputEntity)
          .clear();

        await transactionalEntityManager
          .getRepository(StateEntity)
          .save({ name: l1Monitor, height: this.l1BlockHeight - 1 });
        await transactionalEntityManager
          .getRepository(StateEntity)
          .save({ name: l2Monitor, height: this.l2BlockHeight - 1 });
      }
    );
  }

  // update module name in l2id.move
  updateL2ID() {
    const filePath = path.join(this.contractDir, 'sources', 'l2id.move');
    const fileContent = fs.readFileSync(filePath, 'utf-8');
    const updatedContent = fileContent.replace(
      /(addr::)[^\s{]+( \{)/g,
      `$1${this.moduleName}$2`
    );
    fs.writeFileSync(filePath, updatedContent, 'utf-8');
  }

  publishL2IDMsg(module: string) {
    return new MsgPublish(executor.key.accAddress, [module], 0);
  }

  bridgeInitializeMsg(
    submissionInterval: number,
    finalizedTime: number,
    l2StartBlockHeight: number
  ) {
    return new MsgExecute(
      executor.key.accAddress,
      '0x1',
      'op_bridge',
      'initialize',
      [],
      [
        bcs.serialize('string', this.l2id),
        bcs.serialize('u64', submissionInterval),
        bcs.serialize('address', outputSubmitter.key.accAddress),
        bcs.serialize('address', challenger.key.accAddress),
        bcs.serialize('u64', finalizedTime),
        bcs.serialize('u64', l2StartBlockHeight)
      ]
    );
  }

  bridgeRegisterTokenMsg(metadata: string) {
    return new MsgExecute(
      executor.key.accAddress,
      '0x1',
      'op_bridge',
      'register_token',
      [],
      [bcs.serialize('string', this.l2id), bcs.serialize('object', metadata)]
    );
  }

  async tx(metadata: string) {
    const module = await build(this.contractDir, this.moduleName);
    const msgs = [
      this.publishL2IDMsg(module),
      this.bridgeInitializeMsg(
        this.submissionInterval,
        this.finalizedTime,
        this.l2StartBlockHeight
      ),
      this.bridgeRegisterTokenMsg(metadata)
    ];
    await sendTx(executor, msgs);
  }

  async deployBridge(metadata: string) {
    await initORM();
    const bridge = new Bridge(
      this.submissionInterval,
      this.finalizedTime,
      this.l2StartBlockHeight,
      this.l2id,
      this.contractDir
    );
    await bridge.init();
    await bridge.tx(metadata);
  }
}

export default Bridge;
