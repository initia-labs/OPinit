import { MoveBuilder } from "@initia/builder.js";
// import { AccAddress, MsgPublish } from "@initia/minitia.js";
// import { 
//     Wallet,
//     MnemonicKey,
// } from '@initia/minitia.js';
import { AccAddress, MsgPublish } from "@initia/initia.js";
import { 
    Wallet,
    MnemonicKey,
    LCDClient,
    TxInfo
} from '@initia/initia.js';
import config from '../config';
import { delay } from "bluebird";
async function sendTx(client: LCDClient,sender: Wallet,  msg: any) {
    try {
        const signedTx = await sender.createAndSignTx({msgs:[msg]})
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


async function publishL1(){
    const builder = new MoveBuilder(__dirname+"/L1Contracts",{}); 
    const sender = new Wallet(config.l1lcd, new MnemonicKey({mnemonic: config.EXECUTOR_MNEMONIC}))
    await builder.build();

    const opBatch = await builder.get("op_batch_inbox");
    const bridge = await builder.get("op_bridge");
    const output = await builder.get("op_output");
    const l2id = await builder.get("minitia");
    const publishMsg = new MsgPublish(
        sender.key.accAddress,
        [
            // opBatch.toString("base64"), 
            // bridge.toString("base64"), 
            // output.toString("base64"),
            // l2id.toString("base64")
        ],
        0
    )

    try{
        const txRes = await sendTx(config.l1lcd, sender, publishMsg)
        console.log(txRes)
    }catch (e) {
        console.log(e)
    }
}


// async function publishL2(){
//     const builder = new MoveBuilder(__dirname+"/L2Contracts",{});
//     const sender = new Wallet(config.l2lcd, new MnemonicKey({mnemonic: config.BATCH_SUBMITTER_MNEMONIC}))
//     await builder.build();

//     const bridge = await builder.get("op_bridge");
//     const publishMsg = new MsgPublish(
//         sender.key.accAddress,
//         [bridge.toString("base64")],
//         0
//     )
//     console.log(sender.key.accAddress)
//     console.log(AccAddress.toHex(sender.key.accAddress))
//     try{
//         await transaction(sender, [publishMsg])
//     }catch (e) {
//         console.log(e)
//     }
// }

async function main () {
    await publishL1()
    // await publishL2()
}

main()