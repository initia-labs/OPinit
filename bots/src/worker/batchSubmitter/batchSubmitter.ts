import { getDB } from './db';
import { DataSource, EntityManager } from 'typeorm';
import { batchLogger, batchLogger as logger } from 'lib/logger';
import { BlockBulk, RPCClient } from 'lib/rpc';
import { compressor } from 'lib/compressor';
import { ExecutorOutputEntity, RecordEntity } from 'orm';
import { Wallet, MnemonicKey, MsgRecordBatch } from '@initia/initia.js';
import { delay } from 'bluebird';
import { INTERVAL_BATCH } from 'config';
import { getConfig } from 'config';
import { sendTx } from 'lib/tx';
import MonitorHelper from 'worker/bridgeExecutor/MonitorHelper';

const config = getConfig();

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
        await this.publishBatchToL1(batch);
        await this.saveBatchToDB(
          manager,
          batch,
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

    return compressor(bulk.blocks);
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
  async publishBatchToL1(batch: Buffer) {
    try {
      const executeMsg = new MsgRecordBatch(
        this.submitter.key.accAddress,
        this.bridgeId,
        batch.toString('base64')
      );

      return await sendTx(this.submitter, [executeMsg]);
    } catch (err) {
      throw new Error(`Error publishing batch to L1: ${err}`);
    }
  }

  // Save batch record to database
  async saveBatchToDB(
    manager: EntityManager,
    batch: Buffer,
    batchIndex: number,
    startBlockNumber: number,
    endBlockNumber: number
  ): Promise<RecordEntity> {
    const record = new RecordEntity();

    record.bridgeId = this.bridgeId;
    record.batchIndex = batchIndex;
    record.batch = batch;
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
