import {
  Wallet as mWallet,
  MsgSend,
  Coin,
  MsgExecute
} from '@initia/minitia.js';

import { Wallet as iWallet } from '@initia/initia.js';
import axios from 'axios';
import { getConfig } from 'config';
import { bcs, getOutput, getTx, makeFinalizeMsg } from './helper';
import { sendTx } from 'lib/tx';

const config = getConfig();

export class TxBot {
  l1coin = '0x1::native_uinit::Coin';

  createWallet(lcd, WalletClass, MnemonicKeyClass, mnemonic) {
    return new WalletClass(lcd, new MnemonicKeyClass({ mnemonic }));
  }

  async deposit(sender: iWallet, reciever: mWallet, amount: number) {
    const msg = new MsgExecute(
      sender.key.accAddress,
      '0x1',
      'op_bridge',
      'deposit_token',
      [config.L2ID, this.l1coin],
      [
        bcs.serialize('address', reciever.key.accAddress),
        bcs.serialize('u64', amount)
      ]
    );
    return await sendTx(sender, [msg]);
  }

  async withdrawal(wallet: mWallet, amount: number) {
    const res = await axios.get(`${config.EXECUTOR_URI}/coin/${this.l1coin}`);
    const l2coin = res.data.coin.l2StructTag;

    const msg = new MsgExecute(
      wallet.key.accAddress,
      '0x1',
      'op_bridge',
      'withdraw_token',
      [l2coin],
      [
        bcs.serialize('address', wallet.key.accAddress),
        bcs.serialize('u64', amount)
      ]
    );
    return await sendTx(wallet, [msg]);
  }

  async claim(
    sender: iWallet,
    coinType: string,
    txSequence: number,
    outputIndex: number
  ) {
    const txRes = await getTx(coinType, txSequence);
    const outputRes: any = await getOutput(outputIndex);
    const finalizeMsg = await makeFinalizeMsg(sender, txRes, outputRes.output);
    return await sendTx(sender, [finalizeMsg]);
  }

  // only for L2 account public key gen
  async sendCoin(sender: any, receiver: any, amount: number, denom: string) {
    const msg = new MsgSend(sender.key.accAddress, receiver.key.accAddress, [
      new Coin(denom, amount)
    ]);

    return await sendTx(sender, [msg]);
  }
}
