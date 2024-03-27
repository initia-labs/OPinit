import { MsgProposeOutput } from '@initia/initia.js';
import { INTERVAL_OUTPUT } from 'config';
import { ExecutorOutputEntity } from 'orm';
import { delay } from 'bluebird';
import { outputLogger as logger } from 'lib/logger';
import { ErrorTypes } from 'lib/error';
import { config } from 'config';
import { getBridgeInfo, getLastOutputInfo } from 'lib/query';
import MonitorHelper from 'worker/bridgeExecutor/MonitorHelper';
import { DataSource, EntityManager } from 'typeorm';
import { getDB } from './db';
import { TxWallet, WalletType, getWallet, initWallet } from 'lib/wallet';
import { getCurrentTimeInSecond } from 'lib/util';

export class OutputSubmitter {
  private db: DataSource;
  private submitter: TxWallet;
  private syncedOutputIndex = 1;
  private processedBlockNumber = 1;
  private isRunning = false;
  private bridgeId: number;
  helper: MonitorHelper = new MonitorHelper();

  async init() {
    [this.db] = getDB();
    initWallet(WalletType.OutputSubmitter, config.l1lcd);
    this.submitter = getWallet(WalletType.OutputSubmitter);
    this.bridgeId = config.BRIDGE_ID;
    this.isRunning = true;
  }

  public async run() {
    await this.init();
    await this.validateSubmissionInterval();

    while (this.isRunning) {
      await this.proccessOutput();
    }
  }

  async validateSubmissionInterval() {
    const bridgeInfo = await getBridgeInfo(this.bridgeId);
    const submissionInterval = bridgeInfo.bridge_config.submission_interval;
    const maxSubmissionInterval = Math.floor(submissionInterval.seconds.toNumber() *(2/3));
    if ( config.OUTPUT_PROPOSE_INTERVAL > maxSubmissionInterval ) {
      throw new Error(
        `decrease output propose interval to ${maxSubmissionInterval} seconds`
      );
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
        
        const nextSubmissionTimeSec = (output?.timestamp ?? 0) + config.OUTPUT_PROPOSE_INTERVAL
        if (nextSubmissionTimeSec < getCurrentTimeInSecond()) return
        
        await this.proposeOutput(output);
      });
    } catch (err) {
      if (err.response?.data.type === ErrorTypes.NOT_FOUND_ERROR) {
        logger.info(
          `waiting for output index: ${this.syncedOutputIndex}, processed block number: ${this.processedBlockNumber}`
        );
      } else {
        console.log(err);
        this.stop();
      } 
    } finally {
      await delay(INTERVAL_OUTPUT);
    }
  }

  public async stop() {
    this.isRunning = false;
  }

  private async proposeOutput(outputEntity: ExecutorOutputEntity | null) {
    if (!outputEntity) return;
    const msg = new MsgProposeOutput(
      this.submitter.key.accAddress,
      this.bridgeId,
      outputEntity.endBlockNumber,
      outputEntity.outputRoot
    );

    await this.submitter.transaction([msg]);

    this.processedBlockNumber = outputEntity.endBlockNumber;
    logger.info(
      `successfully submitted! output index: ${this.syncedOutputIndex}, output root: ${outputEntity.outputRoot}`
    );
  }
}
