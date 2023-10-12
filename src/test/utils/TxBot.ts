import {
  Wallet as mWallet,
  MsgSend,
  Coin,
  MsgExecute,
  MnemonicKey as mMnemonicKey
} from '@initia/minitia.js';

import {
  Wallet as iWallet,
  MnemonicKey as iMnemonicKey
} from '@initia/initia.js';
import axios from 'axios';
import { getConfig } from 'config';
import { bcs, executor, getOutput, getTx, makeFinalizeMsg } from './helper';
import { sendTx } from 'lib/tx';
import { computeCoinMetadata, normalizeMetadata } from 'lib/lcd';

const config = getConfig();

export class TxBot {
  l1CoinMetadata: string;
  l1sender: iWallet;
  l2sender: mWallet;
  l1receiver: iWallet;
  l2receiver: mWallet;

  constructor() {
    this.l1CoinMetadata = normalizeMetadata(
      computeCoinMetadata('0x1', 'uinit')
    );
    this.l1sender = this.createWallet(
      config.l1lcd,
      iWallet,
      iMnemonicKey,
      'banner december bunker moral nasty glide slow property pen twist doctor exclude novel top material flee appear imitate cat state domain consider then age'
    );
    this.l2sender = this.createWallet(
      config.l2lcd,
      iWallet,
      iMnemonicKey,
      'banner december bunker moral nasty glide slow property pen twist doctor exclude novel top material flee appear imitate cat state domain consider then age'
    );
    this.l1receiver = this.createWallet(
      config.l1lcd,
      mWallet,
      mMnemonicKey,
      'diamond donkey opinion claw cool harbor tree elegant outer mother claw amount message result leave tank plunge harbor garment purity arrest humble figure endless'
    );
    this.l2receiver = this.createWallet(
      config.l2lcd,
      mWallet,
      mMnemonicKey,
      'diamond donkey opinion claw cool harbor tree elegant outer mother claw amount message result leave tank plunge harbor garment purity arrest humble figure endless'
    );
  }

  createWallet(lcd, WalletClass, MnemonicKeyClass, mnemonic) {
    return new WalletClass(lcd, new MnemonicKeyClass({ mnemonic }));
  }

  async deposit(sender: iWallet, reciever: mWallet, amount: number) {
    const msg = new MsgExecute(
      sender.key.accAddress,
      '0x1',
      'op_bridge',
      'deposit_token',
      [],
      [
        bcs.serialize('address', executor.key.accAddress),
        bcs.serialize('string', config.L2ID),
        bcs.serialize('object', this.l1CoinMetadata),
        bcs.serialize('address', reciever.key.accAddress),
        bcs.serialize('u64', amount)
      ]
    );
    console.log('msg: ', msg);
    return await sendTx(sender, [msg]);
  }

  async withdrawal(wallet: mWallet, amount: number) {
    const res = await axios.get(
      `${config.EXECUTOR_URI}/coin/${this.l1CoinMetadata}`
    );
    const l2CoinMetadata = res.data.coin.l2Metadata;

    const msg = new MsgExecute(
      wallet.key.accAddress,
      '0x1',
      'op_bridge',
      'withdraw_token',
      [],
      [
        bcs.serialize('address', wallet.key.accAddress),
        bcs.serialize('object', l2CoinMetadata),
        bcs.serialize('u64', amount)
      ]
    );
    return await sendTx(wallet, [msg]);
  }

  async claim(sender: iWallet, txSequence: number, outputIndex: number) {
    const txRes = await getTx(this.l1CoinMetadata, txSequence);
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
