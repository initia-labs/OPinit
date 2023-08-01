// import { AccAddress, Coin, MsgDeposit, MsgExecute, MsgPublish, MsgSend } from "@initia/minitia.js";
// import { 
//   Wallet,
//   MnemonicKey,
//   BCS,
//   LCDClient,
//   TxInfo,
// } from '@initia/minitia.js';
import { AccAddress, Coin, MsgExecute, MsgPublish, MsgSend, MsgDeposit } from "@initia/initia.js";
import { 
    Wallet,
    MnemonicKey,
    BCS,
    LCDClient,
    TxInfo,
    Msg
} from '@initia/initia.js';


import { init } from "@sentry/node";
import config from "../config";
import { BridgeConfig } from "lib/types";
import { WalletType, getWallet, initWallet, wallets } from "lib/wallet";
import { delay } from 'bluebird'
import { send } from "process";


const bcs = BCS.getInstance()
const sender = new Wallet(config.l1lcd, new MnemonicKey({mnemonic: config.EXECUTOR_MNEMONIC}))


async function sendTx(client: LCDClient,sender: Wallet,  msg: Msg[]) {
    try {
        const signedTx = await sender.createAndSignTx({msgs:msg})
        const broadcastResult = await client.tx.broadcast(signedTx)
        console.log(broadcastResult)
        await checkTx(client, broadcastResult.txhash)
        return broadcastResult.txhash
    }catch (error) {
        console.log(error)
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


/// outputSubmitter -> op_output/initialize
/// executor -> op_bridge/initialize

async function tx(){
    const from = new Wallet(config.l1lcd, new MnemonicKey({mnemonic: 'stumble much civil surface carry warm suspect print title toe else awake what increase extend island acoustic educate speak viable month brown valve profit'}))
    const to = new Wallet(config.l1lcd, new MnemonicKey({mnemonic: 'file body gasp outside good urban river custom employ supreme ask shoe volcano stamp powder wonder sell balance slab coin mushroom debate funny license'}))

    const challenger = new Wallet(config.l1lcd, new MnemonicKey({mnemonic: config.CHALLENGER_MNEMONIC}))
    const outputSubmitter = new Wallet(config.l1lcd, new MnemonicKey({mnemonic: config.OUTPUT_SUBMITTER_MNEMONIC}))
    // console.log(AccAddress.toHex(sender.key.accAddress))
    const executeMsg = [
      // new MsgExecute(
      //  sender.key.accAddress,
      //  '0x1',
      //   'op_bridge',
      //   'initialize',
      //   [config.L2ID],
      //   []
      // ),
      // new MsgExecute(
      //   sender.key.accAddress,
      //   '0x1',
      //   'op_bridge',
      //   'register_token',
      //   [config.L2ID, '0x1::native_uinit::Coin'],
      //   []
      // ),
      // new MsgExecute(
      //   sender.key.accAddress,
      //   '0x1',
      //   'op_bridge',
      //   'deposit_token',
      //   [config.L2ID, '0x1::native_uinit::Coin'],
      //   [
      //     bcs.serialize('address', sender.key.accAddress),
      //     bcs.serialize('u64', 100)
      //   ]
      // )
      
      // op_output/initialize
      // new MsgExecute(
      //  sender.key.accAddress,
      //  '0x1',
      //   'op_output',
      //   'initialize',
      //   [config.L2ID],
      //   [
      //     bcs.serialize('u64', 100),
      //     bcs.serialize('address', outputSubmitter.key.accAddress),
      //     bcs.serialize('address', challenger.key.accAddress),
      //     bcs.serialize('u64', 3600),
      //     bcs.serialize('u64', 50000)
      //   ]
      // ),

    ]

    await sendTx(config.l1lcd, sender, executeMsg)
}



async function main(){
    await tx()
}

main()