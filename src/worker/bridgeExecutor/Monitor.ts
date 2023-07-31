import * as Bluebird from 'bluebird'
import { RPCSocket } from 'lib/rpc'
import { LCDClient, Event } from '@initia/minitia.js'
import { StateEntity } from 'orm'
import { getDB } from './db'
import { DataSource } from 'typeorm'
import { logger } from 'lib/logger'

export abstract class Monitor {
  public syncedHeight: number
  protected db: DataSource
  protected isRunning = false

  constructor(
    public lcd: LCDClient,
    public socket: RPCSocket
  ) {
    this.db = getDB()[0]
  }

  public async run(): Promise<void> {
    const state = await this.db.getRepository(StateEntity).findOne({
      where: {
        name: this.name(),
      },
    })

    if (!state) {
      await this.db.getRepository(StateEntity).save({ name: this.name(), height: 0 })
    }
    this.syncedHeight = state?.height || 0

    this.socket.initialize()
    this.isRunning = true
    await this.monitor()
  }

  public stop(): void {
    this.socket.stop()
    this.isRunning = false
  }

  public async monitor(): Promise<void> {
    while (this.isRunning) {
      try{
        const latestHeight = this.socket.latestHeight
        if (!latestHeight || this.syncedHeight >= latestHeight) continue
        logger.info(`${this.name()} height ${this.syncedHeight + 1}`)

        const searchRes = await this.lcd.tx.search({
          events: [
            { key: 'tx.height', value: (this.syncedHeight + 1).toString() },
          ],
        })

        const events = searchRes.txs
          .flatMap((tx) => tx.logs ?? [])
          .flatMap((log) => log.events)
        
        await this.handleEvents(events)
        this.syncedHeight += 1
        await this.handleBlock(this.syncedHeight)
        // update state
        await this.db.getRepository(StateEntity).update(
          { name: this.name() },
          { height: this.syncedHeight }
        )
      }catch(e){
        logger.error('Monitor runs error:', e)
      }finally{
        await Bluebird.Promise.delay(100)
      } 
    }
  }

  // eslint-disable-next-line
  public async handleEvents(events: Event[]): Promise<void> {}

  // eslint-disable-next-line
  public async handleBlock(height: number): Promise<void> {}

  // eslint-disable-next-line
  public name(): string {
    return ''
  }
}