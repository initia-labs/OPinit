import { BCS, Msg, MsgExecute } from "@initia/minitia.js";
import { TxWallet, WalletType, getWallet } from 'lib/wallet'
import config from 'config'
import { OutputEntity } from 'orm'
import { APIRequest } from "lib/apiRequest";
const bcs = BCS.getInstance() 

export class OutputSubmitter {
    private submitter: TxWallet;
    private apiRequester: APIRequest;

    async init(){
        this.submitter = getWallet(WalletType.OutputSubmitter)
        this.apiRequester = new APIRequest('https://minitia-executor.initia.tech') // TODO: 
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
            const nextOutputIndex = await this.getNextOutputIndex()
            const outputEntity: OutputEntity = await this.apiRequester.getOuptut(nextOutputIndex) // TODO : get output with startBlockNum
            
            this.proposeL2Output(Buffer.from(outputEntity.outputRoot, 'hex'), nextOutputIndex)
        }catch(err){
            throw new Error(`Error in outputSubmitter: ${err}`)
        }
    }
}