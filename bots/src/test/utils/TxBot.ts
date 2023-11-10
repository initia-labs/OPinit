import {
  Wallet,
  MsgInitiateTokenDeposit,
  Coin,
  MsgInitiateTokenWithdrawal,
  MnemonicKey
} from '@initia/initia.js';
import { makeFinalizeMsg } from './helper';
import { sendTx } from 'lib/tx';
import { getOutputFromExecutor, getWithdrawalTxFromExecutor } from 'lib/query';
import { getConfig } from 'config';

const config = getConfig();

export class TxBot {
  l1sender = new Wallet(
    config.l1lcd,
    new MnemonicKey({
      mnemonic:
        // init1wzenw7r2t2ra39k4l9yqq95pw55ap4sm4vsa9g
        ''
    })
  );

  l1receiver = new Wallet(
    config.l1lcd,
    new MnemonicKey({
      mnemonic:
        // init174knscjg688ddtxj8smyjz073r3w5mmsp3m0m2
        ''
    })
  );

  l2sender = new Wallet(
    config.l2lcd,
    new MnemonicKey({
      mnemonic: ''
    })
  );

  l2receiver = new Wallet(
    config.l2lcd,
    new MnemonicKey({
      mnemonic: ''
    })
  );

  constructor(public bridgeId: number) {}

  async deposit(sender: Wallet, reciever: Wallet, coin: Coin) {
    const msg = new MsgInitiateTokenDeposit(
      sender.key.accAddress,
      this.bridgeId,
      reciever.key.accAddress,
      coin
    );

    return await sendTx(sender, [msg]);
  }

  async withdrawal(sender: Wallet, receiver: Wallet, coin: Coin) {
    const msg = new MsgInitiateTokenWithdrawal(
      sender.key.accAddress,
      receiver.key.accAddress,
      coin
    );

    return await sendTx(sender, [msg]);
  }

  async claim(sender: Wallet, txSequence: number, outputIndex: number) {
    const txRes = await getWithdrawalTxFromExecutor(this.bridgeId, txSequence);
    const outputRes: any = await getOutputFromExecutor(outputIndex);
    const finalizeMsg = await makeFinalizeMsg(
      txRes.withdrawalTx,
      outputRes.output
    );

    const { account_number: accountNumber, sequence } =
      await sender.accountNumberAndSequence();
    return await sendTx(sender, [finalizeMsg], accountNumber, sequence);
  }
}
