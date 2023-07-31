import { MoveBuilder } from "@initia/builder.js";
import { AccAddress, MsgPublish } from "@initia/minitia.js";
import { 
    Wallet,
    MnemonicKey,
} from '@initia/minitia.js';
import config from '../config';
import { transaction } from "../lib/tx";
import { getWallet, initWallet, WalletType } from "../lib/wallet";

async function publishL1(){
    const builder = new MoveBuilder(__dirname+"/L1Contracts",{}); 
    initWallet(WalletType.Executor, config.l1lcd)
    initWallet(WalletType.OutputSubmitter, config.l1lcd)
    const sender =  getWallet(WalletType.OutputSubmitter)
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
            l2id.toString("base64")
        ],
        0
    )

    try{
        const txRes = await sender.transaction([publishMsg])
        console.log(txRes)
    }catch (e) {
        console.log(e)
    }
}


async function publishL2(){
    const builder = new MoveBuilder(__dirname+"/L2Contracts",{});
    const sender = new Wallet(config.l2lcd, new MnemonicKey({mnemonic: config.BATCH_SUBMITTER_MNEMONIC}))
    await builder.build();

    const bridge = await builder.get("op_bridge");
    const publishMsg = new MsgPublish(
        sender.key.accAddress,
        [bridge.toString("base64")],
        0
    )
    console.log(sender.key.accAddress)
    console.log(AccAddress.toHex(sender.key.accAddress))
    try{
        await transaction(sender, [publishMsg])
    }catch (e) {
        console.log(e)
    }
}

async function main () {
    await publishL1()
    // await publishL2()
}

main()