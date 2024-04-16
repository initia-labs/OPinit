import { MsgProposeOutput } from '@initia/initia.js'
import { INTERVAL_OUTPUT } from '../../config'
import { ExecutorOutputEntity } from '../../orm'
import { delay } from 'bluebird'
import { outputLogger as logger } from '../../lib/logger'
import { ErrorTypes } from '../../lib/error'
import { config } from '../../config'
import { getLastOutputInfo } from '../../lib/query'
import MonitorHelper from '../../lib/monitor/helper'
import { DataSource, EntityManager } from 'typeorm'
import { getDB } from './db'
import { TxWallet, WalletType, getWallet, initWallet } from '../../lib/wallet'

export class OutputSubmitter {
  private db: DataSource
  private submitter: TxWallet
  private syncedOutputIndex = 1
  private processedBlockNumber = 1
  private isRunning = false
  private bridgeId: number
  helper: MonitorHelper = new MonitorHelper()

  async init() {
    [this.db] = getDB()
    initWallet(WalletType.OutputSubmitter, config.l1lcd)
    this.submitter = getWallet(WalletType.OutputSubmitter)
    this.bridgeId = config.BRIDGE_ID
    this.isRunning = true
  }

  public async run() {
    await this.init()

    while (this.isRunning) {
      await this.processOutput()
    }
  }

  async processOutput() {
    try {
      await this.db.transaction(async (manager: EntityManager) => {
        const lastOutputInfo = await getLastOutputInfo(this.bridgeId)
        if (lastOutputInfo) {
          this.syncedOutputIndex = lastOutputInfo.output_index + 1
        }

        const output = await this.helper.getOutputByIndex(
          manager,
          ExecutorOutputEntity,
          this.syncedOutputIndex
        )
        if (!output) return

        await this.proposeOutput(output)
        logger.info(
          `successfully submitted! output index: ${this.syncedOutputIndex}, output root: ${output.outputRoot} (${output.startBlockNumber}, ${output.endBlockNumber})`
        )
      })
    } catch (err) {
      if (err.response?.data.type === ErrorTypes.NOT_FOUND_ERROR) {
        logger.info(
          `waiting for output index: ${this.syncedOutputIndex}, processed block number: ${this.processedBlockNumber}`
        )
        await delay(INTERVAL_OUTPUT)
      } else {
        logger.error(`Output Submitter halted! ${err}`)
        this.stop()
      }
    }
  }

  public async stop() {
    this.isRunning = false
  }

  private async proposeOutput(outputEntity: ExecutorOutputEntity) {
    const msg = new MsgProposeOutput(
      this.submitter.key.accAddress,
      this.bridgeId,
      outputEntity.endBlockNumber,
      outputEntity.outputRoot
    )

    await this.submitter.transaction([msg])

    this.processedBlockNumber = outputEntity.endBlockNumber
  }
}
