import { Wallet, MnemonicKey, MsgProposeOutput } from '@initia/initia.js';
import { INTERVAL_OUTPUT } from 'config';
import { ExecutorOutputEntity } from 'orm';
import { delay } from 'bluebird';
import { outputLogger as logger } from 'lib/logger';
import { ErrorTypes } from 'lib/error';
import { getConfig } from 'config';
import { sendTx } from 'lib/tx';
import { getLastOutputInfo } from 'lib/query';
import MonitorHelper from 'worker/bridgeExecutor/MonitorHelper';
import { DataSource, EntityManager } from 'typeorm';
import { getDB } from './db';

const config = getConfig();

export class OutputSubmitter {
  private db: DataSource;
  private submitter: Wallet;
  private syncedOutputIndex = 1;
  private processedBlockNumber = 1;
  private isRunning = false;
  private bridgeId: number;
  helper: MonitorHelper = new MonitorHelper();

  async init() {
    [this.db] = getDB();
    this.submitter = new Wallet(
      config.l1lcd,
      new MnemonicKey({ mnemonic: config.OUTPUT_SUBMITTER_MNEMONIC })
    );
    this.bridgeId = config.BRIDGE_ID;
    this.isRunning = true;
  }

  public async run() {
    await this.init();

    while (this.isRunning) {
      await this.proccessOutput();
    }
  }

  async proccessOutput() {
    try {
      await this.db.transaction(async (manager: EntityManager) => {
        const lastOutputInfo = await getLastOutputInfo(this.bridgeId);
        if (lastOutputInfo) {
          this.syncedOutputIndex = lastOutputInfo.output_index + 1;
        }

        const output = await this.helper.getOutputByIndex(
          manager,
          ExecutorOutputEntity,
          this.syncedOutputIndex
        );
        if (!output) return;

        await this.proposeOutput(output);
        logger.info(
          `successfully submitted! output index: ${this.syncedOutputIndex}, output root: ${output.outputRoot}`
        );
      });
    } catch (err) {
      if (err.response?.data.type === ErrorTypes.NOT_FOUND_ERROR) {
        logger.warn(
          `waiting for output index: ${this.syncedOutputIndex}, processed block number: ${this.processedBlockNumber}`
        );
        await delay(INTERVAL_OUTPUT);
      } else {
        logger.error(err);
        this.stop();
      }
    }
  }

  public async stop() {
    this.isRunning = false;
  }

  private async proposeOutput(outputEntity: ExecutorOutputEntity) {
    const msg = new MsgProposeOutput(
      this.submitter.key.accAddress,
      this.bridgeId,
      outputEntity.endBlockNumber,
      outputEntity.outputRoot
    );

    const { account_number, sequence } =
      await this.submitter.accountNumberAndSequence();

    await sendTx(this.submitter, [msg], account_number, sequence);

    this.processedBlockNumber = outputEntity.endBlockNumber;
  }
}
