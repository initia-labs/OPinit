import config from "../../config"
import axios from 'axios';
import { 
    Wallet,
    MnemonicKey,
    MsgExecute,
    BCS,
    LCDClient,
    TxInfo,
    Msg
} from '@initia/initia.js';
import { DataSource } from "typeorm"
import { WithdrawalTx, L2TokenBridgeInitiatedEvent, L1TokenBridgeInitiatedEvent } from "lib/types"
import { WithdrawalStorage } from "lib/storage"
import { createOutputRoot } from "lib/util";
import { getDB } from "./db";
import { delay } from 'bluebird'
import { logger } from "lib/logger";

import { StateEntity, WithdrawalTxEntity, DepositTxEntity } from "orm";
import { OutputEntity, TxEntity } from "orm";
import { fetchBridgeConfig } from "lib/lcd";
const bcs = BCS.getInstance()

export class Challenger{
    private challenger: Wallet;
    private challengeDone: boolean;
    private db: DataSource;

    private L1SyncedHeight: number
    private L2SyncedHeight: number
    private L1MonitorName = 'l1_challenger_monitor'
    private L2MonitorName = 'l2_challenger_monitor'
    private submissionInterval: number;


    async init() {
        [this.db] = getDB();    
        this.challenger = new Wallet(config.l1lcd, new MnemonicKey({mnemonic: config.CHALLENGER_MNEMONIC}));
        this.challengeDone = false

        const bridgeCfg = await fetchBridgeConfig()
        this.submissionInterval = parseInt(bridgeCfg.submission_interval)
        console.log(this.submissionInterval)
    }

    async getSyncedHeight(monitorName: string, syncedHeightProperty: string) {
        const state = await this.db.getRepository(StateEntity).findOne({
            where: { name: monitorName },
        });
        if (!state) {
            await this.db.getRepository(StateEntity).save({ name: monitorName, height: 0 });
        }
        this[syncedHeightProperty] = state?.height || 0;
    }
    
    public async getL1SyncedHeight() {
        await this.getSyncedHeight(this.L1MonitorName, 'L1SyncedHeight');
    }
    
    public async getL2SyncedHeight() {
        await this.getSyncedHeight(this.L2MonitorName, 'L2SyncedHeight');
    }
    
    private async monitorTransactions(
        monitorName: string,
        syncedHeightProperty: string,
        getTxEvents: (height: number) => Promise<any[] | null>,
        buildTxEntities: (txs: any[], height: number) => any[],
        txEntityRepository: any
    ) {
        for (;;) {
            await this[`get${syncedHeightProperty}`]();
            this[syncedHeightProperty] += 1;
    
            await this.db.getRepository(StateEntity).update(
                { name: monitorName },
                { height: this[syncedHeightProperty] }
            );
    
            const txs: any[] | null = await getTxEvents(this[syncedHeightProperty]);
            if (!txs) return;
    
            const txEntities = buildTxEntities(txs, this[syncedHeightProperty]);
            this.db.getRepository(txEntityRepository).save(txEntities);
        }
    }
    
    public async monitorL1Deposit() {
        await this.monitorTransactions(
            this.L1MonitorName,
            'L1SyncedHeight',
            this.getDepositTxEvents,
            (txs, height) => txs.map((tx) => ({
                height: height,
                from: tx.from,
                to: tx.to,
                l2_id: tx.l2_id,
                l1_token: tx.l1_token,
                l2_token: tx.l2_token.toString('hex'),
                amount: tx.amount,
                l1_sequence: tx.l1_sequence,
                is_checked: false
            })),
            DepositTxEntity
        );
    }
    
    public async monitorL2Withdrawal() {
        await this.monitorTransactions(
            this.L2MonitorName,
            'L2SyncedHeight',
            this.getWithdrawalTxEvents,
            (txs, height) => txs.map((tx) => ({ 
                height: height,
                from: tx.from,
                to: tx.to,
                l2_token: tx.l2_token.toString('hex'),
                amount: tx.amount,
                l2_sequence: tx.l2_sequence,
                is_checked: false
             })),
            WithdrawalTxEntity
        );
    }

    
    // monitoring L1 deposit event and check the relayer works properly (L1 TokenBridgeInitiatedEvent)
    public async l1Challenge(){
        
        return
    }

    // monitoring L2 withdrawal event and check the relayer works properly (L2 TokenBridgeInitiatedEvent)
    public async l2Challenge(){
        // if (!await this.isReadyToChallenge()) return

        // const startHeight = this.processedHeight + 1
        // const outputIndex = this.processedOutputIndex + 1
        // const nextHeight = startHeight + this.submissionInterval
        
        // const [storage, withdrawalTxs] = await this.makeWithdrawalStorage(startHeight, nextHeight-1)
        // const outputRootL1 = await this.getL1OutputRoot(outputIndex)
        // const outputRootL2 = await this.makeL2OutputRoot(outputIndex, nextHeight-1, storage)
        
        // if (outputRootL1 !== outputRootL2) {
        //     this.doChallenge(outputIndex)
        //     return
        // }

        // await this.saveWithdrawalTxs(storage, withdrawalTxs, outputIndex)
        // // await this.saveOutputRoot(outputIndex, outputRootL2, nextHeight-1)
        
        // this.processedHeight = nextHeight
        // this.processedOutputIndex = outputIndex
    }

    // // check whether there's challengable tx
    // public async isReadyToChallenge(): Promise<boolean> {
    //     const outputEntity = await this.db.getRepository(OutputEntity).find({
    //         order: { outputIndex: 'DESC'},
    //         take: 1
    //     })
    //     this.processedOutputIndex = outputEntity[0]?.outputIndex ?? -1
    //     this.processedHeight = outputEntity[0]?.startBlockHeight ?? 0

    //     const lastHeight = await getLatestBlockHeight(config.L2_RPC_URI)
    //     return lastHeight >= (this.processedOutputIndex + 1) * this.submissionInterval + this.l2StartHeight - 1
    // }

    async getL1OutputRoot(outputIndex: number): Promise<string> {
        const outputRoot = await config.l1lcd.move.viewFunction<Buffer>(
            '0x1',
            'op_output',
            'get_output_root',
            [config.L2ID],
            [
                bcs.serialize('u64', outputIndex)
            ]
        )
        return outputRoot.toString('hex')
    }

    async makeL2OutputRoot(outputIndex: number, lastL2BlockHeight: number, storage: WithdrawalStorage): Promise<string>{ 
        const res = await axios.get(`${config.L2_LCD_URI}/cosmos/base/tendermint/v1beta1/blocks/${lastL2BlockHeight}`)

        const version = outputIndex 
        const stateRoot= res['data']['block']['header']['app_hash']// app hash
        const storageRoot= storage.getMerkleRoot() // storage 
        const lastBlockHash= res['data']['block_id']['hash'] // block hash
        const challengerOutputRoot = createOutputRoot(version, stateRoot, storageRoot, lastBlockHash)
        return challengerOutputRoot
    }

    // monitor L1 deposit events
    public async getDepositTxEvents(height: number): Promise<L1TokenBridgeInitiatedEvent[]|null>{
        const res = await axios.get(`${config.L1_LCD_URI}/cosmos/tx/v1beta1/txs?events=tx.height=${height}`)
        
        const evtName = `0x1::op_bridge::TokenBridgeInitiatedEvent`
        const txResponses = res['data']['tx_responses']
        if (!txResponses) return null
        const depositTxs : L1TokenBridgeInitiatedEvent[] = []
        txResponses.forEach(logs => {
            logs.events.forEach(event => {
                if (event.type !== 'move') return
 
                for (const attr of event.attributes) {
                    if (attr.key === 'type_tag' && attr.value === evtName ) {
                        const dataAttr = event.attributes.find(attr => attr.key === 'data');
                        if (!dataAttr) {
                            return null
                        }
                        const events: L1TokenBridgeInitiatedEvent = JSON.parse(dataAttr.value)
                        
                        depositTxs.push(events)
                    }
                }
            })
        })
        return depositTxs
    }

    // monitor L2 withdrawal events
    public async getWithdrawalTxEvents(height: number): Promise<L2TokenBridgeInitiatedEvent[]|null>{
        const res = await axios.get(`${config.L2_LCD_URI}/cosmos/tx/v1beta1/txs?events=tx.height=${height}`)
        
        const evtName = `0x1::op_bridge::TokenBridgeInitiatedEvent`
        const txResponses = res['data']['tx_responses']
        if (!txResponses) return null
        const withdrawalTxs : L2TokenBridgeInitiatedEvent[] = []
        txResponses.forEach(logs => {
            logs.events.forEach(event => {
                if (event.type !== 'move') return
 
                for (const attr of event.attributes) {
                    if (attr.key === 'type_tag' && attr.value === evtName ) {
                        const dataAttr = event.attributes.find(attr => attr.key === 'data');
                        if (!dataAttr) {
                            return null
                        }
                        const events: L2TokenBridgeInitiatedEvent = JSON.parse(dataAttr.value)

                        withdrawalTxs.push(events)
                    }
                }
            })
        })
        return withdrawalTxs
    }
    
    // save withdrawal txs in [height, height + submission_interval)
    async saveWithdrawalTxs(storage: WithdrawalStorage, withdrawalTxs: WithdrawalTx[], outputIndex: number){
        const storageRoot = storage.getMerkleRoot()

        // save merkle root and proof for each tx
        for (const withdrawalTx of withdrawalTxs) {
            const txEntity: TxEntity = {
                sequence: withdrawalTx.sequence,
                sender: withdrawalTx.sender,
                receiver: withdrawalTx.receiver,
                amount: withdrawalTx.amount,
                coin_type: withdrawalTx.coin_type,
                outputIndex: outputIndex,
                merkleRoot: storageRoot,
                merkleProof: storage.getMerkleProof(withdrawalTx),
               
            }
            await this.db.getRepository(TxEntity).save(txEntity)
        }
    }
    
    async saveOutputRoot(
        outputIndex: number, 
        outputRoot: string, 
        stateRoot: string, 
        storageRoot: string, 
        lastBlockHash: string, 
        startBlockHeight: number
    ){
        const outputEntity: OutputEntity = {
            outputIndex,
            outputRoot,
            stateRoot,
            storageRoot,
            lastBlockHash,
            startBlockHeight,
        }

        await this.db.getRepository(OutputEntity).save(outputEntity)
    }

    // // save withdrawal txs in [start, end)
    // async makeWithdrawalStorage(start: number, end: number): Promise<[WithdrawalStorage, WithdrawalTx[]]>{
    //     const withdrawalTxs: WithdrawalTx[] = []
        
    //     for (let i = start; i <= end; i++) {
    //         const txEntity = await this.getWithdrawalTxEvents(i)
    //         if (txEntity === null) continue
    //         withdrawalTxs.push(...txEntity)
    //     }

    //     const storage = new WithdrawalStorage(withdrawalTxs)
    //     return [storage, withdrawalTxs]
    // }

    // // delete invalid l2 output 
    // public async doChallenge(outputIndex: number) {
    //     const executeMsg = new MsgExecute(
    //         this.challenger.key.accAddress,
    //         '0x1',
    //         'op_output',
    //         'delete_l2_output',
    //         [config.L2ID],
    //         [bcs.serialize('u64', outputIndex)]
    //     )

    //     // await transaction(this.challenger, [executeMsg])
    //     await sendTx(config.l1lcd, this.challenger, [executeMsg])
    //     this.challengeDone = true
    // }

    // public isChallengeDone(): boolean {
    //     return this.challengeDone
    // }
}

// /// Utils
// async function sendTx(client: LCDClient,sender: Wallet,  msg: Msg[]) {
//     try {
//         const signedTx = await sender.createAndSignTx({msgs:msg})
//         const broadcastResult = await client.tx.broadcast(signedTx)
//         await checkTx(client, broadcastResult.txhash)
//         return broadcastResult.txhash
//     }catch (error) {
//         console.log(error)
//         throw new Error(`Error in sendTx: ${error}`)
//     }
// }

// export async function checkTx(
//     lcd: LCDClient,
//     txHash: string,
//     timeout = 60000
// ): Promise<TxInfo | undefined> {
//     const startedAt = Date.now()

//     while (Date.now() - startedAt < timeout) {
//         try {
//         const txInfo = await lcd.tx.txInfo(txHash)
//         if (txInfo) return txInfo
//         await delay(1000)
//         } catch (err) {
//         throw new Error(`Failed to check transaction status: ${err.message}`)
//         }
//     }

//     throw new Error('Transaction checking timed out');
// }