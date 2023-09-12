import { MsgExecute, MsgPublish } from '@initia/initia.js';
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
  tag: string;
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
    this.tag = this.l2id.split('::')[1];
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

  updateL2ID() {
    const filePath = path.join(this.contractDir, 'sources', 'l2id.move');
    const fileContent = fs.readFileSync(filePath, 'utf-8');
    const updatedContent = fileContent.replace(
      /(addr::)[^\s{]+( \{)/g,
      `$1${this.tag}$2`
    );
    fs.writeFileSync(filePath, updatedContent, 'utf-8');
  }

  async publishL2ID(dirname: string, moduleName: string) {
    const sender = executor;
    const module = await build(dirname, moduleName);
    const executeMsg = [new MsgPublish(sender.key.accAddress, [module], 0)];
    await sendTx(sender, executeMsg);
  }

  async bridgeInitialize() {
    const sender = executor;
    const executeMsg = [
      new MsgExecute(
        sender.key.accAddress,
        '0x1',
        'op_bridge',
        'initialize',
        [this.l2id],
        []
      )
    ];
    await sendTx(sender, executeMsg);
  }

  async outputInitialize(
    submissionInterval: number,
    finalizedTime: number,
    l2StartBlockHeight: number
  ) {
    const sender = executor;
    const executeMsg = [
      new MsgExecute(
        sender.key.accAddress,
        '0x1',
        'op_output',
        'initialize',
        [this.l2id],
        [
          bcs.serialize('u64', submissionInterval),
          bcs.serialize('address', outputSubmitter.key.accAddress),
          bcs.serialize('address', challenger.key.accAddress),
          bcs.serialize('u64', finalizedTime),
          bcs.serialize('u64', l2StartBlockHeight)
        ]
      )
    ];
    await sendTx(sender, executeMsg);
  }

  async bridgeRegisterToken(coinType: string) {
    const sender = executor;
    const executeMsg = [
      new MsgExecute(
        sender.key.accAddress,
        '0x1',
        'op_bridge',
        'register_token',
        [this.l2id, coinType],
        []
      )
    ];
    await sendTx(sender, executeMsg);
  }

  async tx() {
    await this.publishL2ID(this.contractDir, this.tag);
    console.log('publish L2ID done');

    await this.bridgeInitialize();
    console.log('initialize bridge done');

    await this.outputInitialize(
      this.submissionInterval,
      this.finalizedTime,
      this.l2StartBlockHeight
    );
    console.log('output initiaization done');
    
    await this.bridgeRegisterToken(`0x1::native_uinit::Coin`);
    console.log('register token done');
  }

  async deployBridge() {
    await initORM();
    const bridge = new Bridge(
      this.submissionInterval,
      this.finalizedTime,
      this.l2StartBlockHeight,
      this.l2id,
      this.contractDir
    );
    await bridge.init();
    await bridge.tx();
  }
}

export default Bridge;
