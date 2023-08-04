import config from '../../config';
import axios from 'axios';
import {
  Wallet,
  MnemonicKey,
  BCS,
  LCDClient,
  Msg,
  TxInfo,
  MsgExecute
} from '@initia/initia.js';
import { DataSource } from 'typeorm';
import {
  WithdrawalTx,
  L2TokenBridgeInitiatedEvent,
  L1TokenBridgeInitiatedEvent,
  DepositTx
} from 'lib/types';
import { WithdrawalStorage } from 'lib/storage';
import { createOutputRoot } from 'lib/util';
import { getDB } from './db';
import { getL2Denom, structTagToDenom } from 'lib/util';
import {
  StateEntity,
  WithdrawalTxEntity,
  DepositTxEntity,
  ChallengerOutputEntity
} from 'orm';
import { fetchBridgeConfig } from 'lib/lcd';
import { delay } from 'bluebird';
import { logger } from 'lib/logger';

const bcs = BCS.getInstance();

export class Challenger {
  private challenger: Wallet
  private isRunning = false;
  private db: DataSource;

  private submissionInterval: number;

  async init() {
    [this.db] = getDB();
    this.challenger = new Wallet(
      config.l1lcd,
      new MnemonicKey({ mnemonic: config.CHALLENGER_MNEMONIC })
    );

    const bridgeCfg = await fetchBridgeConfig();
    this.submissionInterval = parseInt(bridgeCfg.submission_interval);
  }

  public async run(): Promise<void> {
    while (!this.isRunning) {
      await this.l1Challenge();
      await this.l2Challenge();
    }
  }

  // monitoring L1 deposit event and check the relayer works properly (L1 TokenBridgeInitiatedEvent)
  public async l1Challenge() {
    return;
  }

  public stop(): void {
    this.isRunning = false;
  }

  // get unchecked withdrawal txs with smallest sequence from db
  async getUncheckedWitdrawalTx(): Promise<WithdrawalTxEntity | null> {
    const uncheckedWithdrawalTx: WithdrawalTxEntity[] = await this.db
      .getRepository(WithdrawalTxEntity)
      .find({
        where: { isChecked: false },
        order: { sequence: 'ASC' },
        take: 1
      });

    if (uncheckedWithdrawalTx.length === 0) return null;
    return uncheckedWithdrawalTx[0];
  }

  async checkWithdrawalTx(
    withdrawalTxEntity: WithdrawalTxEntity
  ): Promise<void> {
    await this.db
      .getRepository(WithdrawalTxEntity)
      .update(withdrawalTxEntity.outputIndex, { isChecked: true });
  }

  async getChallengerOutputRoot(outputIndex: number): Promise<string | null> {
    const challengerOutputEntity = await this.db
      .getRepository(ChallengerOutputEntity)
      .find({
        where: { outputIndex: outputIndex },
        take: 1
      });

    if (challengerOutputEntity.length === 0) return null;
    return challengerOutputEntity[0].outputRoot;
  }

  async getContractOutputRoot(outputIndex: number): Promise<string | null> {
    const outputRootFromContract = await config.l1lcd.move.viewFunction<Buffer>(
      '0x1',
      'op_output',
      'get_output_root',
      [config.L2ID],
      [bcs.serialize(BCS.U64, outputIndex)]
    );
    return outputRootFromContract.toString('hex');
  }

  // monitoring L2 withdrawal event and check the relayer works properly (L2 TokenBridgeInitiatedEvent)
  public async l2Challenge() {
    const uncheckedWithdrawalTx = await this.getUncheckedWitdrawalTx();
    if (uncheckedWithdrawalTx === null) {
      logger.info('[L2 Challenger] no unchecked withdrawal tx');
      return;
    }

    const outputRootFromContract = await this.getContractOutputRoot(
      uncheckedWithdrawalTx.outputIndex
    );
    if (outputRootFromContract === null) {
      logger.info(
        `[L2 Challenger] contract output root not found for output index ${uncheckedWithdrawalTx.outputIndex}`
      );
      return;
    }

    const outputRootFromChallenger = await this.getChallengerOutputRoot(
      uncheckedWithdrawalTx.outputIndex
    );
    if (outputRootFromChallenger === null) {
      logger.info(
        `[L2 Challenger] challenger output root not found for output index ${uncheckedWithdrawalTx.outputIndex}`
      );
      return;
    }

    if (outputRootFromContract === outputRootFromChallenger) {
      logger.info(
        `[L2 Challenger] output root matched for output index ${uncheckedWithdrawalTx.outputIndex}`
      );
      await this.checkWithdrawalTx(uncheckedWithdrawalTx);
      return;
    }

    await this.deleteL2Ouptut(uncheckedWithdrawalTx.outputIndex);
  }

  async deleteL2Ouptut(outputIndex: number) {
    const executeMsg: Msg = new MsgExecute(
      this.challenger.key.accAddress,
      '0x1',
      'op_output',
      'delete_l2_output',
      [config.L2ID],
      [bcs.serialize(BCS.U64, outputIndex)]
    );

    await sendTx(config.l1lcd, this.challenger, [executeMsg]);
    this.isRunning = true;
  }
}

/// Utils
async function sendTx(client: LCDClient, sender: Wallet, msg: Msg[]) {
  try {
    const signedTx = await sender.createAndSignTx({ msgs: msg });
    const broadcastResult = await client.tx.broadcast(signedTx);
    await checkTx(client, broadcastResult.txhash);
    return broadcastResult.txhash;
  } catch (error) {
    console.log(error);
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
