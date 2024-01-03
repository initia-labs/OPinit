import { getDB } from './db';
import FailedTxEntity from 'orm/executor/FailedTxEntity';
import { Coin, MsgFinalizeTokenDeposit } from '@initia/initia.js';
import { INTERVAL_MONITOR, config } from 'config';
import { DataSource } from 'typeorm';
import Bluebird from 'bluebird';
import winston from 'winston';
import { TxWallet, WalletType, getWallet, initWallet } from 'lib/wallet';
import { buildFailedTxNotification, notifySlack } from 'lib/slack';

export class Resurrector {
  private db: DataSource;
  isRunning = true;
  executor: TxWallet;
  errorCounter = 0;

  constructor(public logger: winston.Logger) {
    [this.db] = getDB();
    initWallet(WalletType.Executor, config.l2lcd);
    this.executor = getWallet(WalletType.Executor);
  }

  async updateProcessed(failedTx: FailedTxEntity): Promise<void> {
    await this.db.getRepository(FailedTxEntity).update(
      {
        bridgeId: failedTx.bridgeId,
        sequence: failedTx.sequence,
        processed: false
      },
      { processed: true }
    );

    this.logger.info(
      `Resurrected failed tx: ${failedTx.bridgeId} ${failedTx.sequence}`
    );
  }

  async resubmitFailedDepositTx(failedTx: FailedTxEntity): Promise<void> {
    const msg = new MsgFinalizeTokenDeposit(
      this.executor.key.accAddress,
      failedTx.sender,
      failedTx.receiver,
      new Coin(failedTx.l2Denom, failedTx.amount),
      parseInt(failedTx.sequence),
      failedTx.l1Height,
      Buffer.from(failedTx.data, 'hex').toString('base64')
    );
    try {
      await this.executor.transaction([msg]);
      await this.updateProcessed(failedTx);
    } catch (err) {
      if (this.errorCounter++ < 20) {
        await Bluebird.delay(5 * 1000);
        return;
      }
      this.errorCounter = 0;
      await notifySlack(buildFailedTxNotification(failedTx));
    }
  }

  async getFailedTxs(): Promise<FailedTxEntity[]> {
    return await this.db.getRepository(FailedTxEntity).find({
      where: {
        processed: false
      }
    });
  }

  public async ressurect(): Promise<void> {
    const failedTxs = await this.getFailedTxs();

    for (const failedTx of failedTxs) {
      const error = failedTx.error;

      // Check x/opchild/errors.go
      if (error.includes('deposit already finalized')) {
        await this.updateProcessed(failedTx);
        continue;
      }
      await this.resubmitFailedDepositTx(failedTx);
    }
  }

  stop(): void {
    this.isRunning = false;
  }

  public async run() {
    while (this.isRunning) {
      try {
        await this.ressurect();
      } catch (err) {
        this.stop();
        throw new Error(err);
      } finally {
        await Bluebird.delay(INTERVAL_MONITOR);
      }
    }
  }
}
