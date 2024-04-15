import Bluebird from 'bluebird'
import { RPCClient, RPCSocket } from '../rpc'
import { StateEntity } from '../../orm'
import { DataSource, EntityManager } from 'typeorm'
import MonitorHelper from './helper'
import winston from 'winston'
import { INTERVAL_MONITOR, config } from '../../config'

const MAX_BLOCKS = 20 // DO NOT CHANGE THIS, hard limit is 20 in cometbft.
const MAX_RETRY_INTERVAL = 30_000

export abstract class Monitor {
  public syncedHeight: number
  public currentHeight: number
  protected db: DataSource
  protected isRunning = false
  protected bridgeId: number
  protected retryNum = 0
  helper: MonitorHelper = new MonitorHelper()

  constructor(
    public socket: RPCSocket,
    public rpcClient: RPCClient,
    public logger: winston.Logger
  ) {
    this.bridgeId = config.BRIDGE_ID
  }

  public async run(): Promise<void> {
    const state = await this.db.getRepository(StateEntity).findOne({
      where: {
        name: this.name()
      }
    })

    this.syncedHeight = state?.height || 0

    if (!state) {
      if (this.name() === 'executor_l1_monitor') {
        this.syncedHeight = config.EXECUTOR_L1_MONITOR_HEIGHT
      } else if (this.name() === 'executor_l2_monitor') {
        this.syncedHeight = config.EXECUTOR_L2_MONITOR_HEIGHT
      }

      await this.db
        .getRepository(StateEntity)
        .save({ name: this.name(), height: this.syncedHeight })
    }

    this.socket.initialize()
    this.isRunning = true
    await this.monitor()
  }

  public stop(): void {
    this.socket.stop()
    this.isRunning = false
  }

  async handleBlockWithStateUpdate(manager: EntityManager): Promise<void> {
    await this.handleBlock(manager)
    if (this.syncedHeight % 10 === 0) {
      this.logger.info(`${this.name()} height ${this.syncedHeight}`)
    }
    this.syncedHeight++
    await manager
      .getRepository(StateEntity)
      .update({ name: this.name() }, { height: this.syncedHeight })
  }

  public async monitor(): Promise<void> {
    while (this.isRunning) {
      try {
        const latestHeight = this.socket.latestHeight
        if (!latestHeight || !(latestHeight > this.syncedHeight)) continue
        const blockchainData = await this.rpcClient.getBlockchain(
          this.syncedHeight + 1,
          // cap the query to fetch 20 blocks at maximum
          // DO NOT CHANGE THIS, hard limit is 20 in cometbft.
          Math.min(latestHeight, this.syncedHeight + MAX_BLOCKS)
        )
        if (blockchainData === null) continue

        await this.db.transaction(async (manager: EntityManager) => {
          for (const metadata of blockchainData.block_metas.reverse()) {
            this.currentHeight = this.syncedHeight + 1

            if (this.currentHeight !== parseInt(metadata.header.height)) {
              throw new Error(
                `expected block meta is the height ${this.currentHeight}, but got ${metadata.header.height}`
              )
            }

            if (parseInt(metadata.num_txs) === 0) {
              await this.handleBlockWithStateUpdate(manager)
              continue
            }

            // handle event always called when there is a tx in a block,
            // so empty means, the tx indexing is still on going.
            const ok: boolean = await this.handleEvents(manager)
            if (!ok) {
              this.retryNum++
              if (this.retryNum * INTERVAL_MONITOR >= MAX_RETRY_INTERVAL) {
                // rotate when tx index data is not found during 30s after block stored.
                this.rpcClient.rotateRPC()
              }
              break
            }
            this.retryNum = 0
            await this.handleBlockWithStateUpdate(manager)
          }
        })
      } catch (err) {
        this.logger.info(err)
        this.stop()
        throw new Error(`Error in ${this.name()} ${err}`)
      } finally {
        await Bluebird.delay(INTERVAL_MONITOR)
      }
    }
  }

  // eslint-disable-next-line
  public async handleEvents(manager: EntityManager): Promise<any> {}

  // eslint-disable-next-line
  public async handleBlock(manager: EntityManager): Promise<void> {}

  // eslint-disable-next-line
  public name(): string {
    return ''
  }
}
