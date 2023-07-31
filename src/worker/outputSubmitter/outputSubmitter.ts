import { BCS, Msg, MsgExecute, Wallet } from "@initia/minitia.js";
import { createOutputRoot } from "../../lib/util";
import { TxWallet, WalletType, getWallet, wallets } from '../../lib/wallet'
import config from '../../config'
import { DataSource } from "typeorm"
import { getDB } from "./db";

const bcs = BCS.getInstance() 

export class OutputSubmitter {
    private outputIndex = 0;
    private dataSource: DataSource;
    private submitter: TxWallet;

    async init(){
        [this.dataSource] = getDB()
        this.submitter = getWallet(WalletType.OutputSubmitter)
    }
    async isFinalized(outputIndex: number){
        return await config.l1lcd.move.viewFunction<boolean>(
            '0x1',
            'op_output',
            'is_finalized',
            [config.L2ID],
            [
                bcs.serialize('u64', outputIndex)
            ])
    }

    async proposeL2Output(outputRoot: Buffer, l2BlockHeight: number){
        const executeMsg: Msg = new MsgExecute(
            this.submitter.key.accAddress,
            '0x1',
            'op_output',
            'propose_l2_output',
            [config.L2ID],
            [
                bcs.serialize('vector<u8>', outputRoot, 10000),
                bcs.serialize('u64', l2BlockHeight)
            ]
        )
        this.submitter.transaction([executeMsg])
    }

    public async run(){
        try{
            if (await this.isFinalized(this.outputIndex)) {
                this.outputIndex += 1 // get from db
                return
            }

            const outputRoot = createOutputRoot(
                this.outputIndex,
                '',
                '',
                ''
            )
            
            this.proposeL2Output(Buffer.from(outputRoot), this.outputIndex)
        }catch(err){
            throw new Error(`Error in outputSubmitter: ${err}`)
        }
    }
}