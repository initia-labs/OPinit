import { BCS, Msg, MsgExecute, Wallet, MnemonicKey, LCDClient, TxInfo } from "@initia/initia.js";
import config from 'config'
import { OutputEntity } from 'orm'
import { APIRequest } from "lib/apiRequest";
import { delay } from 'bluebird'
import { logger } from "lib/logger";
const bcs = BCS.getInstance() 

export class OutputSubmitter {
    private submitter: Wallet;
    private apiRequester: APIRequest;
    private syncedOutputIndex: number;

    async init(){
        this.submitter = new Wallet(config.l1lcd, new MnemonicKey({ mnemonic: config.OUTPUT_SUBMITTER_MNEMONIC}))
        this.apiRequester = new APIRequest(config.EXECUTOR_URI)
        this.syncedOutputIndex = -1
    }

    async getNextOutputIndex(){
        return await config.l1lcd.move.viewFunction<number>(
            '0x1',
            'op_output',
            'next_output_index',
            [config.L2ID],
            []
        )
    }

    async getNextBlockHeight(){
        return await config.l1lcd.move.viewFunction<number>(
            '0x1',
            'op_output',
            'next_block_num',
            [config.L2ID],
            []
        )
    }

    async proposeL2Output(outputRoot: Buffer, l2BlockHeight: number){
        const executeMsg: Msg = new MsgExecute(
            this.submitter.key.accAddress,
            '0x1',
            'op_output',
            'propose_l2_output',
            [config.L2ID],
            [
                bcs.serialize('vector<u8>', outputRoot, 33),
                bcs.serialize('u64', l2BlockHeight)
            ]
        )
        await sendTx(config.l1lcd, this.submitter, [executeMsg])
    }

    public async run(){
        try{
            const nextOutputIndex = await this.getNextOutputIndex()
            const nextBlockHeight = await this.getNextBlockHeight()
            // logger.info(`current index: ${this.syncedOutputIndex} next index: ${nextOutputIndex}`)
            logger.info(`next block height ${nextBlockHeight}`)

            if (nextOutputIndex <= this.syncedOutputIndex) {
                logger.info(`waiting for next output index.`)
                return
            }
            const outputEntity: OutputEntity = await this.apiRequester.getOuptut(nextOutputIndex)
            
            await this.proposeL2Output(Buffer.from(outputEntity.outputRoot, 'hex'), nextBlockHeight)
            this.syncedOutputIndex = nextOutputIndex
            logger.info(`submitted output index: ${nextOutputIndex} output root: ${outputEntity.outputRoot}`)
        }catch(err){
            throw new Error(`Error in outputSubmitter: ${err}`)
        } finally {
            await delay(1000)
        } 
    }
}

/// Utils
async function sendTx(client: LCDClient,sender: Wallet,  msg: Msg[]) {
    try {
        const signedTx = await sender.createAndSignTx({msgs:msg})
        const broadcastResult = await client.tx.broadcast(signedTx)
        await checkTx(client, broadcastResult.txhash)
        return broadcastResult.txhash
    }catch (error) {
        throw new Error(`Error in sendTx: ${error}`)
    }
}

export async function checkTx(
    lcd: LCDClient,
    txHash: string,
    timeout = 60000
): Promise<TxInfo | undefined> {
    const startedAt = Date.now()

    while (Date.now() - startedAt < timeout) {
        try {
        const txInfo = await lcd.tx.txInfo(txHash)
        if (txInfo) return txInfo
        await delay(1000)
        } catch (err) {
        throw new Error(`Failed to check transaction status: ${err.message}`)
        }
    }

    throw new Error('Transaction checking timed out');
}
