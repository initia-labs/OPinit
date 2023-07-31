import { AccAddress, Coin, MsgDeposit, MsgExecute, MsgPublish, MsgSend } from "@initia/minitia.js";
import { 
    Wallet,
    MnemonicKey,
    BCS,
    LCDClient,
    TxInfo,
} from '@initia/minitia.js';
import { init } from "@sentry/node";
import config from "config";
import { BridgeConfig } from "lib/types";
import { WalletType, getWallet, initWallet, wallets } from "lib/wallet";

const bcs = BCS.getInstance()
const L1Client= config.l1lcd
const L2Client= config.l2lcd
const sender = new Wallet( L1Client, new MnemonicKey({mnemonic: 'recycle sight world spoon leopard shine dizzy before public use jungle either arctic detail hawk output option august hedgehog menu keen night work become'}))
const bridgeAddr = '0x1'
const L2ID = '0x56ccf33c45b99546cd1da172cf6849395bbf8573::l2::Id'


async function sendTx(client: LCDClient,sender: Wallet,  msg: any) {
    try {
        const signedTx = await sender.createAndSignTx({msgs:[msg]})
        const broadcastResult = await client.tx.broadcast(signedTx)
        await pollingTx(client, broadcastResult.txhash)
        return broadcastResult.txhash
    }catch (error) {
        console.log(error)
        throw new Error(`Error in sendTx: ${error}`)
    }
}

async function pollingTx(lcd: LCDClient, txhash : string): Promise<void>{
    return new Promise((resolve, reject) => {
        const polling = setInterval(async () => {
            let txResult: TxInfo | null = null;
            try {
                txResult = await lcd.tx.txInfo(txhash)        
            }catch (error){
                reject(error)
                clearInterval(polling);
            }

            if (txResult) {
                resolve();
                clearInterval(polling);
            }
            
        }, 1000);
    })
}


/// outputSubmitter -> op_output/initialize
/// executor -> op_bridge/initialize

async function tx(){
    initWallet(WalletType.Executor, config.l1lcd)
    // initWallet(WalletType.OutputSubmitter, config.l1lcd)
    // initWallet(WalletType.Challenger, config.l1lcd)
    const executor = getWallet(WalletType.Executor)
    const sender = new Wallet(config.l1lcd, new MnemonicKey({mnemonic: 'stumble much civil surface carry warm suspect print title toe else awake what increase extend island acoustic educate speak viable month brown valve profit'}))
    const receiver = new Wallet(config.l2lcd, new MnemonicKey({mnemonic: 'file body gasp outside good urban river custom employ supreme ask shoe volcano stamp powder wonder sell balance slab coin mushroom debate funny license'}))
    console.log(executor.key.accAddress)
            // 'initialize',
        // [config.L2ID],
        // []
        // 'register_token',
        // [config.L2ID, '0x1::native_uinit::Coin'],
        // []
    const executeMsg = new MsgDeposit(
        sender.key.accAddress,
        sender.key.accAddress,
        receiver.key.accAddress,
        new Coin('uinit', 100),
        0
    )

    await sendTx(L1Client, executor, executeMsg)

    // const outputSubmitter = getWallet(WalletType.OutputSubmitter)
    // const challenger = getWallet(WalletType.Challenger)
    // const executeMsg = new MsgExecute(
    //     outputSubmitter.key.accAddress,
    //     '0x1',
    //     'op_output',
    //     'initialize',
    //     [config.L2ID],
    //     [
    //         bcs.serialize('u64', 100),
    //         bcs.serialize('address', outputSubmitter.key.accAddress),
    //         bcs.serialize('address', challenger.key.accAddress),
    //         bcs.serialize('u64', 3600),
    //         bcs.serialize('u64', 50000)
    //     ])
    // await sendTx(L1Client, outputSubmitter, executeMsg)

    // const res = await config.l1lcd.move
    //     .viewFunction<{ submission_interval: string }>(
    //         '0x1',
    //         'op_output',
    //         'get_config_store',
    //         [config.L2ID],
    //         []
    //     )
    //     .then((res) => [
    //         Number.parseInt(res['submission_interval']),
    //         Number.parseInt(res['starting_block_number']),
    //     ])
    //     .catch((err) => {
    //         console.log(err)
    //     })
    // console.log(res)
}



async function main(){
    await tx()
}

main()