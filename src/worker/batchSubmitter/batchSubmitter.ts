import { initORM, getDB } from './db'
import { DataSource } from "typeorm"
import { logger } from "../../lib/logger"
import config from "../../config"
import { BlockBulk, getBlockBulk } from '../../lib/rpc';
import { compressor } from '../../lib/compressor';
import { transaction, getLatestBlockHeight } from '../../lib/tx';
import { RecordEntity } from '../../orm';
import * as bluebird  from 'bluebird';
import L2Monitoring from '../l2Monitoring';
import { 
    Wallet,
    MnemonicKey,
    MsgExecute,
    BCS,
} from '@initia/minitia.js';
import { fetchBridgeConfig } from 'lib/lcd';

const bcs = BCS.getInstance() 

export class BatchSubmitter {
    private batchIndex = 0;
    private batchL2StartHeight: number;
    private latestBlockHeight: number;
    private dataSource: DataSource;
    private submitter: Wallet;
    private submissionInterval: number;

    async init() {
        [this.dataSource] = getDB()
        this.latestBlockHeight = await getLatestBlockHeight(config.l2lcd);
        this.submitter = new Wallet(config.l1lcd, new MnemonicKey({mnemonic: config.BATCH_SUBMITTER_MNEMONIC}))
        const bridgeCfg = await fetchBridgeConfig()
        this.batchL2StartHeight = parseInt(bridgeCfg.starting_block_number)
        this.submissionInterval = parseInt(bridgeCfg.submission_interval)
    }

    public async run() {
        try{
            const latestBatch = await this.getStoredBatch(this.dataSource);
            if (latestBatch) {
                this.batchIndex = latestBatch.batchIndex + 1;
            }

            // e.g [start_height + 0, start_height + 99], [start_height + 100, start_height + 199], ...
            const startHeight = this.batchL2StartHeight + this.batchIndex * this.submissionInterval;
            const endHeight = this.batchL2StartHeight + (this.batchIndex + 1) * this.submissionInterval - 1;

            if (endHeight > this.latestBlockHeight) {
                logger.info(`[${this.batchIndex}th batch] batch interval is not satisfied. current height: ${this.latestBlockHeight} target height: ${endHeight}`);
                this.latestBlockHeight = await getLatestBlockHeight(config.l2lcd); // update latest block height
                return
            }

            const batch = await this.getBatch(startHeight, endHeight);
            logger.info(`[${this.batchIndex}th batch] batch is generated. start height: ${startHeight} end height: ${endHeight}`);
            const txHash = await this.publishBatchToL1(batch);
            logger.info(`[${this.batchIndex}th batch] batch is published to L2. tx hash: ${txHash}`);

            await this.saveBatchToDB(this.dataSource, batch, this.batchIndex);
            logger.info(`[${this.batchIndex}th batch] batch is indexed to DB`);
        } catch (err) {
            throw new Error(`Error in BatchSubmitter: ${err}`);
        }
    }

    // Get [start, end] batch from L2 
    async getBatch(start:number, end:number): Promise<Buffer>{ 
        const bulk: BlockBulk| null = (await getBlockBulk(start.toString(), end.toString()))
        if (!bulk) {
            throw new Error(`Error getting block bulk from L2`)
        }
        const batch = compressor(bulk.blocks)
        return batch
    }

    async getStoredBatch(db: DataSource): Promise<RecordEntity | null> {
        const storedRecord = await db.getRepository(RecordEntity)
            .find({
                order: {
                    batchIndex: "DESC"
                },
                take: 1
            })
            .catch((err) => {
                logger.error(`Error getting stored batch: ${err}`)
                return null
            })
    
        return storedRecord ? storedRecord[0] : null
    }

    // Publish a batch to L1
    async publishBatchToL1(batch : Buffer){
        try{
            const executeMsg = new MsgExecute(
                this.submitter.key.accAddress,
                '0x1',
                'op_batch_inbox',
                'record_batch',
                [config.L2ID],
                [
                    bcs.serialize('vector<u8>',batch, 100000) // TODO: get max batch size from chain
                ]
            )
            transaction(this.submitter, [executeMsg])
        } catch (err){
            throw new Error(`Error publishing batch to L1: ${err}`)
        }
    }

    // Save batch record to database
    async saveBatchToDB(db: DataSource, batch: Buffer, batchIndex: number): Promise<RecordEntity>{   
        const record = new RecordEntity()
        
        record.l2Id = config.L2ID
        record.batchIndex = batchIndex
        record.batch = batch

        await db.getRepository(RecordEntity)
            .save(record)
            .catch((error) => {
                throw new Error(`Error saving record ${record.l2Id} batch ${batchIndex} to database: ${error}`)
            })
        
        return record
    }
}
