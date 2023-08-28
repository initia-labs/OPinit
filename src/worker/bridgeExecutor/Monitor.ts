import * as Bluebird from 'bluebird';
import { RPCSocket } from 'lib/rpc';
import { StateEntity } from 'orm';
import { DataSource } from 'typeorm';
import MonitorHelper from './MonitorHelper';
import winston from 'winston';
import { INTERVAL_MONITOR } from 'config';

export abstract class Monitor {
  public syncedHeight: number;
  protected db: DataSource;
  protected isRunning = false;
  helper: MonitorHelper = new MonitorHelper();
  constructor(public socket: RPCSocket, public logger: winston.Logger) {}

  public async run(): Promise<void> {
    const state = await this.db.getRepository(StateEntity).findOne({
      where: {
        name: this.name()
      }
    });

    if (!state) {
      await this.db
        .getRepository(StateEntity)
        .save({ name: this.name(), height: 0 });
    }
    this.syncedHeight = state?.height || 0;

    this.socket.initialize();
    this.isRunning = true;
    await this.monitor();
  }

  public stop(): void {
    this.socket.stop();
    this.isRunning = false;
  }

  public async monitor(): Promise<void> {
    while (this.isRunning) {
      try {
        const latestHeight = this.socket.latestHeight;
        if (!latestHeight || this.syncedHeight >= latestHeight) continue;
        if ((this.syncedHeight + 1) % 10 == 0 && this.syncedHeight !== 0) {
          this.logger.info(`${this.name()} height ${this.syncedHeight + 1}`);
        }
        await this.handleEvents();

        this.syncedHeight += 1;
        await this.handleBlock();

        // update state
        await this.db
          .getRepository(StateEntity)
          .update({ name: this.name() }, { height: this.syncedHeight });
      } catch (err) {
        this.stop();
        throw new Error(`Error in ${this.name()} ${err}`)
      } finally {
        await Bluebird.Promise.delay(INTERVAL_MONITOR);
      }
    }
  }

  // eslint-disable-next-line
  public async handleEvents(): Promise<void> {}

  // eslint-disable-next-line
  public async handleBlock(): Promise<void> {}

  // eslint-disable-next-line
  public name(): string {
    return '';
  }
}
