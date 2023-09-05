import {
  AccAddress,
  Coin,
  MsgExecute,
  MsgPublish,
  MsgSend,
  MsgDeposit
} from '@initia/initia.js';
import {
  Wallet,
  MnemonicKey,
  BCS,
  LCDClient,
  TxInfo,
  Msg
} from '@initia/initia.js';
import axios from 'axios';
import * as fs from 'fs';
import * as path from 'path';
import config from 'config';
import { delay } from 'bluebird';
import { MoveBuilder } from '@initia/builder.js';
import { getDB, initORM } from 'worker/bridgeExecutor/db';
import { DataSource, EntityManager } from 'typeorm';
import {
  ExecutorCoinEntity,
  ExecutorOutputEntity,
  StateEntity,
  ExecutorWithdrawalTxEntity
} from 'orm';

export const bcs = BCS.getInstance();

export async function sendTx(client: LCDClient, sender: Wallet, msg: Msg[]) {
  try {
    const signedTx = await sender.createAndSignTx({ msgs: msg });
    const broadcastResult = await client.tx.broadcast(signedTx);
    console.log('height: ', broadcastResult.height);
    await checkTx(client, broadcastResult.txhash);
    return broadcastResult.txhash;
  } catch (error) {
    console.log(error?.response?.data);
    throw new Error(`Error in sendTx: ${error}`);
  }
}

export async function checkTx(
  lcd: LCDClient,
  txHash: string,
  timeout = 60000
): Promise<TxInfo | undefined> {
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeout) {
    try {
      const txInfo = await lcd.tx.txInfo(txHash);
      if (txInfo) return txInfo;
      await delay(1000);
    } catch (err) {
      throw new Error(`Failed to check transaction status: ${err.message}`);
    }
  }

  throw new Error('Transaction checking timed out');
}

/// outputSubmitter -> op_output/initialize
/// executor -> op_bridge/initialize
export async function build(dirname: string, moduleName: string) {
  const builder = new MoveBuilder(__dirname + `/${dirname}`, {});
  await builder.build();
  const contract = await builder.get(moduleName);
  return contract.toString('base64');
}

export const executor = new Wallet(
  config.l1lcd,
  new MnemonicKey({ mnemonic: config.EXECUTOR_MNEMONIC })
);
export const challenger = new Wallet(
  config.l1lcd,
  new MnemonicKey({ mnemonic: config.CHALLENGER_MNEMONIC })
);
export const outputSubmitter = new Wallet(
  config.l1lcd,
  new MnemonicKey({ mnemonic: config.OUTPUT_SUBMITTER_MNEMONIC })
);

class Bridge {
  db: DataSource;
  submissionInterval: number;
  finalizedTime: number;
  l1BlockHeight: number;
  l2BlockHeight: number;
  l2id: string;
  tag: string;
  contractDir: string;

  constructor(
    submissionInterval: number,
    finalizedTime: number,
    l2id: string,
    contractDir: string
  ) {
    [this.db] = getDB();
    this.submissionInterval = submissionInterval;
    this.finalizedTime = finalizedTime;
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
    const filePath = path.join(
      __dirname,
      this.contractDir,
      'sources',
      'l2id.move'
    );
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
    await sendTx(config.l1lcd, sender, executeMsg);
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
    await sendTx(config.l1lcd, sender, executeMsg);
  }

  async outputInitialize(
    submissionInterval: number,
    finalizedTime: number,
    l2BlockHeight: number
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
          bcs.serialize('u64', l2BlockHeight)
        ]
      )
    ];
    await sendTx(config.l1lcd, sender, executeMsg);
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
    await sendTx(config.l1lcd, sender, executeMsg);
  }

  async tx() {
    await this.publishL2ID(this.contractDir, this.tag);
    await delay(7000);
    console.log('publish L2ID done');

    await this.bridgeInitialize();
    await delay(7000);
    console.log('initialize bridge done');

    await this.outputInitialize(
      this.submissionInterval,
      this.finalizedTime,
      this.l2BlockHeight
    );
    await delay(7000);
    console.log('setup bridge and output done');

    await this.bridgeRegisterToken(`0x1::native_uinit::Coin`);
    console.log('register token done');
    await delay(7000);
    console.log(
      `L1 Block Height: ${this.l1BlockHeight}, L2 Block Height: ${this.l2BlockHeight}, L2ID: ${this.l2id}`
    );
    console.log(
      `submissionInterval: ${this.submissionInterval}, finalizedTime: ${this.finalizedTime}`
    );
  }

  async deployBridge() {
    await initORM();
    const bridge = new Bridge(
      this.submissionInterval,
      this.finalizedTime,
      this.l2id,
      this.contractDir
    );
    await bridge.init();
    await bridge.tx();
  }
}

export default Bridge;
