import { getDB } from './db'
import UnconfirmedTxEntity from '../../orm/executor/UnconfirmedTxEntity'
import { Coin, MsgFinalizeTokenDeposit } from '@initia/initia.js'
import { INTERVAL_MONITOR, config } from '../../config'
import { DataSource } from 'typeorm'
import Bluebird from 'bluebird'
import winston from 'winston'
import { TxWallet, WalletType, getWallet, initWallet } from '../../lib/wallet'
import { buildFailedTxNotification, buildResolveErrorNotification, notifySlack } from '../../lib/slack'

export class Resurrector {
  private db: DataSource
  isRunning = true
  executor: TxWallet
  errorCounter = 0

  constructor(public logger: winston.Logger) {
    [this.db] = getDB()
    initWallet(WalletType.Executor, config.l2lcd)
    this.executor = getWallet(WalletType.Executor)
  }

  async updateProcessed(unconfirmedTx: UnconfirmedTxEntity): Promise<void> {
    await this.db.getRepository(UnconfirmedTxEntity).update(
      {
        bridgeId: unconfirmedTx.bridgeId,
        sequence: unconfirmedTx.sequence,
        processed: false
      },
      { processed: true }
    )

    this.logger.info(
      `Resurrected failed tx: ${unconfirmedTx.bridgeId} ${unconfirmedTx.sequence}`
    )
  }

  async resubmitFailedDepositTx(unconfirmedTx: UnconfirmedTxEntity): Promise<void> {
    const txKey = `${unconfirmedTx.sender}-${unconfirmedTx.receiver}-${unconfirmedTx.amount}`
    const msg = new MsgFinalizeTokenDeposit(
      this.executor.key.accAddress,
      unconfirmedTx.sender,
      unconfirmedTx.receiver,
      new Coin(unconfirmedTx.l2Denom, unconfirmedTx.amount),
      parseInt(unconfirmedTx.sequence),
      unconfirmedTx.l1Height,
      unconfirmedTx.l1Denom,
      Buffer.from(unconfirmedTx.data, 'hex').toString('base64')
    )
    try {
      await this.executor.transaction([msg])
      await this.updateProcessed(unconfirmedTx)
      await notifySlack(txKey, buildResolveErrorNotification(`[INFO] Transaction successfully resubmitted and processed for ${unconfirmedTx.sender} to ${unconfirmedTx.receiver} of amount ${unconfirmedTx.amount}.`), false)
    } catch (err) {
      if (this.errorCounter++ < 20) {
        await Bluebird.delay(5 * 1000)
        return
      }
      this.errorCounter = 0
      await notifySlack(txKey, buildFailedTxNotification(unconfirmedTx))
    }
  }

  async getunconfirmedTxs(): Promise<UnconfirmedTxEntity[]> {
    return await this.db.getRepository(UnconfirmedTxEntity).find({
      where: {
        processed: false
      }
    })
  }

  public async ressurect(): Promise<void> {
    const unconfirmedTxs = await this.getunconfirmedTxs()

    for (const unconfirmedTx of unconfirmedTxs) {
      const error = unconfirmedTx.error

      // Check x/opchild/errors.go
      if (error.includes('deposit already finalized')) {
        await this.updateProcessed(unconfirmedTx)
        continue
      }
      await this.resubmitFailedDepositTx(unconfirmedTx)
    }
  }

  stop(): void {
    this.isRunning = false
  }

  public async run() {
    while (this.isRunning) {
      try {
        await this.ressurect()
      } catch (err) {
        this.stop()
        throw new Error(err)
      } finally {
        await Bluebird.delay(INTERVAL_MONITOR)
      }
    }
  }
}
