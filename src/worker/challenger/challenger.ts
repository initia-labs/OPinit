import config from "../../config"
import axios from 'axios';
import { 
    Wallet,
    MnemonicKey,
    MsgExecute,
    BCS,
} from '@initia/minitia.js';
import { transaction } from "../../lib/tx";
import { getLatestBlockHeight } from "../../lib/rpc";
import { DataSource } from "typeorm"
import { OutputEntity, TxEntity } from 'orm'
import { WithdrawalTx } from "lib/types"
import { WithdrawalStorage } from "lib/storage"
import { createOutputRoot } from "lib/util";
import { fetchBridgeConfig } from 'lib/lcd'
import { getDB } from "./db";

const bcs = BCS.getInstance()

export class Challenger{
    private challenger: Wallet;
    private challengeDone: boolean;
    private processedHeight: number; // most recently challenged height (at most 7 days ago)
    private processedOutputIndex: number; // most recently challenged output_index
    private dataSource: DataSource;
    private submissionInterval: number;
    private l2StartHeight: number

    async init() {
        [this.dataSource] = getDB();    
        this.challenger = new Wallet(config.l1lcd, new MnemonicKey({mnemonic: 'recycle sight world spoon leopard shine dizzy before public use jungle either arctic detail hawk output option august hedgehog menu keen night work become'}));
        this.challengeDone = false
        const bridgeCfg = await fetchBridgeConfig()
        this.submissionInterval= parseInt(bridgeCfg.submission_interval)
        this.l2StartHeight = parseInt(bridgeCfg.starting_block_number)
    }

    // monitoring L1 deposit event and check the relayer works properly (L1 TokenBridgeInitiatedEvent)
    public async l1Challenge(){
        return
    }

    // monitoring L2 withdrawal event and check the relayer works properly (L2 TokenBridgeInitiatedEvent)
    public async l2Challenge(){
        if (!await this.isReadyToChallenge()) return

        const startHeight = this.processedHeight + 1
        const outputIndex = this.processedOutputIndex + 1
        const nextHeight = startHeight + this.submissionInterval
        
        const [storage, withdrawalTxs] = await this.makeWithdrawalStorage(startHeight, nextHeight-1)
        const outputRootL1 = await this.getL1OutputRoot(outputIndex)
        const outputRootL2 = await this.makeL2OutputRoot(outputIndex, nextHeight-1, storage)
        
        if (outputRootL1 !== outputRootL2) {
            this.doChallenge(outputIndex)
            return
        }

        await this.saveWithdrawalTxs(storage, withdrawalTxs, outputIndex)
        // await this.saveOutputRoot(outputIndex, outputRootL2, nextHeight-1)
        
        this.processedHeight = nextHeight
        this.processedOutputIndex = outputIndex
    }


    public async run(){
        await this.l1Challenge()
        await this.l2Challenge()
    }
    
    // check whether there's challengable tx
    public async isReadyToChallenge(): Promise<boolean> {
        const outputEntity = await this.dataSource.getRepository(OutputEntity).find({
            order: { outputIndex: 'DESC'},
            take: 1
        })
        this.processedOutputIndex = outputEntity[0]?.outputIndex ?? -1
        this.processedHeight = outputEntity[0]?.startBlockHeight ?? 0

        const lastHeight = await getLatestBlockHeight(config.L2_RPC_URI)
        return lastHeight >= (this.processedOutputIndex + 1) * this.submissionInterval + this.l2StartHeight - 1
    }

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

        const version = outputIndex // TODO: is it right? what if output is deleted?
        const stateRoot= res['data']['block']['header']['app_hash']// app hash
        const storageRoot= storage.getMerkleRoot() // storage 
        const lastBlockHash= res['data']['block_id']['hash'] // block hash
        const challengerOutputRoot = createOutputRoot(version, stateRoot, storageRoot, lastBlockHash)
        return challengerOutputRoot
    }

    public async getDepositTx(height: number): Promise<WithdrawalTx[]|null>{
        const res = await axios.get(`${config.L1_LCD_URI}/cosmos/tx/v1beta1/txs?events=tx.height=${height}`)
        
        const evtName = `0x1::op_bridge::TokenBridgeInitiatedEvent`
        const txResponses = res['data']['tx_responses']
        if (!txResponses) return null
        const withdrawalTxs : WithdrawalTx[] = []
        txResponses.forEach(logs => {
            logs.events.forEach(event => {
                if (event.type !== 'move') return
 
                for (const attr of event.attributes) {
                    if (attr.key === 'type_tag' && attr.value === evtName ) {
                        const dataAttr = event.attributes.find(attr => attr.key === 'data');
                        if (!dataAttr) {
                            return null
                        }
                        const withdrawalTx: WithdrawalTx = JSON.parse(dataAttr.value)
                        
                        withdrawalTxs.push(withdrawalTx)
                    }
                }
            })
        })
        return withdrawalTxs
    }


    public async getWithdrawalTx(height: number): Promise<WithdrawalTx[]|null>{
        const res = await axios.get(`${config.L2_LCD_URI}/cosmos/tx/v1beta1/txs?events=tx.height=${height}`)
        
        const evtName = `0x1::op_bridge::TokenBridgeInitiatedEvent`
        const txResponses = res['data']['tx_responses']
        if (!txResponses) return null
        const withdrawalTxs : WithdrawalTx[] = []
        txResponses.forEach(logs => {
            logs.events.forEach(event => {
                if (event.type !== 'move') return
 
                for (const attr of event.attributes) {
                    if (attr.key === 'type_tag' && attr.value === evtName ) {
                        const dataAttr = event.attributes.find(attr => attr.key === 'data');
                        if (!dataAttr) {
                            return null
                        }
                        const withdrawalTx: WithdrawalTx = JSON.parse(dataAttr.value)
                        
                        withdrawalTxs.push(withdrawalTx)
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
            await this.dataSource.getRepository(TxEntity).save(txEntity)
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

        await this.dataSource.getRepository(OutputEntity).save(outputEntity)
    }

    // save withdrawal txs in [start, end)
    async makeWithdrawalStorage(start: number, end: number): Promise<[WithdrawalStorage, WithdrawalTx[]]>{
        const withdrawalTxs: WithdrawalTx[] = []
        
        for (let i = start; i <= end; i++) {
            const txEntity = await this.getWithdrawalTx(i)
            if (txEntity === null) continue
            withdrawalTxs.push(...txEntity)
        }

        const storage = new WithdrawalStorage(withdrawalTxs)
        return [storage, withdrawalTxs]
    }

    // delete invalid l2 output 
    public async doChallenge(outputIndex: number) {
        const executeMsg = new MsgExecute(
            this.challenger.key.accAddress,
            '0x1',
            'op_output',
            'delete_l2_output',
            [config.L2ID],
            [bcs.serialize('u64', outputIndex)]
        )

        await transaction(this.challenger, [executeMsg])
        this.challengeDone = true
    }

    public isChallengeDone(): boolean {
        return this.challengeDone
    }
}