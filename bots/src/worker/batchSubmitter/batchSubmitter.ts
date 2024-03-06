import { getDB } from './db';
import { DataSource, EntityManager } from 'typeorm';
import { batchLogger, batchLogger as logger } from 'lib/logger';
import { BlockBulk, RPCClient } from 'lib/rpc';
import { compress } from 'lib/compressor';
import { ExecutorOutputEntity, RecordEntity } from 'orm';
import {
  Wallet,
  MnemonicKey,
  MsgRecordBatch,
  Fee,
  Coins
} from '@initia/initia.js';
import { delay } from 'bluebird';
import { INTERVAL_BATCH } from 'config';
import { config } from 'config';
import { sendTx } from 'lib/tx';
import MonitorHelper from 'worker/bridgeExecutor/MonitorHelper';
import { safeSubmitPayForBlob } from 'celestia/utils';

export class BatchSubmitter {
  private batchIndex = 0;
  private db: DataSource;
  private submitter: Wallet;
  private bridgeId: number;
  private isRunning = false;
  private rpcClient: RPCClient;
  helper: MonitorHelper = new MonitorHelper();

  async init() {
    [this.db] = getDB();
    this.rpcClient = new RPCClient(config.L2_RPC_URI, batchLogger);
    this.submitter = new Wallet(
      config.l1lcd,
      new MnemonicKey({ mnemonic: config.BATCH_SUBMITTER_MNEMONIC })
    );

    this.bridgeId = config.BRIDGE_ID;
    this.isRunning = true;
  }

  public stop() {
    this.isRunning = false;
  }

  public async run() {
    await this.init();

    while (this.isRunning) {
      await this.processBatch();
    }
  }

  async processBatch() {
    try {
      await this.db.transaction(async (manager: EntityManager) => {
        const latestBatch = await this.getStoredBatch(manager);
        this.batchIndex = latestBatch ? latestBatch.batchIndex + 1 : 1;
        const output = await this.helper.getOutputByIndex(
          manager,
          ExecutorOutputEntity,
          this.batchIndex
        );

        if (!output) return;

        const batch = await this.getBatch(
          output.startBlockNumber,
          output.endBlockNumber
        );
        let batchInfo: string[];

        if (config.PUBLISH_BATCH_TARGET === 'l1') {
          batchInfo = await this.publishBatchToL1(batch);
        } else if (config.PUBLISH_BATCH_TARGET === 'celestia') {
          batchInfo = await this.publishBatchToCelestia(batch);
        } else {
          throw Error('Unknown publish batch target');
        }
        await this.saveBatchToDB(
          manager,
          batchInfo,
          this.batchIndex,
          output.startBlockNumber,
          output.endBlockNumber
        );
        logger.info(`${this.batchIndex}th batch is successfully saved`);
      });
    } catch (err) {
      throw new Error(`Error in BatchSubmitter: ${err}`);
    } finally {
      await delay(INTERVAL_BATCH);
    }
  }

  // Get [start, end] batch from L2
  async getBatch(start: number, end: number): Promise<Buffer> {
    const bulk: BlockBulk | null = await this.rpcClient.getBlockBulk(
      start.toString(),
      end.toString()
    );
    if (!bulk) {
      throw new Error(`Error getting block bulk from L2`);
    }

    return compress(bulk.blocks);
  }

  async getStoredBatch(manager: EntityManager): Promise<RecordEntity | null> {
    const storedRecord = await manager.getRepository(RecordEntity).find({
      order: {
        batchIndex: 'DESC'
      },
      take: 1
    });

    return storedRecord[0] ?? null;
  }

  // Publish a batch to L1
  async publishBatchToL1(batch: Buffer): Promise<string[]> {
    try {
      const base = 200000;
      const perByte = 10;
      const maxBytes = 500000; // 500kb

      const batchInfos: string[] = [];

      while (batch.length !== 0) {
        let subData: Buffer;
        if (batch.length > maxBytes) {
          subData = batch.slice(0, maxBytes);
          batch = batch.slice(maxBytes);
        } else {
          subData = batch;
          batch = Buffer.from([]);
        }
        const executeMsg = new MsgRecordBatch(
          this.submitter.key.accAddress,
          this.bridgeId,
          subData.toString('base64')
        );

        const gasLimit = Math.floor((base + perByte * subData.length) * 1.2);
        const fee = getFee(this.submitter, gasLimit);

        const batchInfo = await sendTx(this.submitter, [executeMsg], fee);
        batchInfos.push(batchInfo.txhash);

        await delay(1000); // break for each tx ended
      }

      return batchInfos;
    } catch (err) {
      throw new Error(`Error publishing batch to L1: ${err}`);
    }
  }

  // Publish a batch to Celestia
  async publishBatchToCelestia(batch: Buffer): Promise<string[]> {
    try {
      return await safeSubmitPayForBlob(batch);
    } catch (err) {
      throw new Error(`Error publishing batch to celestia: ${err}`);
    }
  }

  // Save batch record to database
  async saveBatchToDB(
    manager: EntityManager,
    batchInfo: string[],
    batchIndex: number,
    startBlockNumber: number,
    endBlockNumber: number
  ): Promise<RecordEntity> {
    const record = new RecordEntity();

    record.bridgeId = this.bridgeId;
    record.batchIndex = batchIndex;
    record.batchInfo = batchInfo;
    record.startBlockNumber = startBlockNumber;
    record.endBlockNumber = endBlockNumber;

    await manager
      .getRepository(RecordEntity)
      .save(record)
      .catch((error) => {
        throw new Error(
          `Error saving record ${record.bridgeId} batch ${batchIndex} to database: ${error}`
        );
      });

    return record;
  }
}

function getFee(wallet: Wallet, gasLimit: number): Fee {
  const gasPrices = new Coins(wallet.lcd.config.gasPrices).toArray();
  if (gasPrices.length === 0) {
    throw Error('gasPrices must be set');
  }
  const gasPrice = gasPrices[0];
  const gasAmount = gasPrice.mul(gasLimit).toIntCeilCoin();

  const fee = new Fee(gasLimit, [gasAmount]);
  return fee;
}
