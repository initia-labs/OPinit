import { Coin, MnemonicKey, MsgSend, Wallet } from '@initia/initia.js';
import { delay } from 'bluebird';
import { config } from 'config';
import { sendTx } from 'lib/tx';
import { TxBot } from 'test/utils/TxBot';

const L1_FUNDER = new Wallet(
  config.l1lcd,
  new MnemonicKey({
    mnemonic: ''
  })
);
const L2_FUNDER = new Wallet(
  config.l2lcd,
  new MnemonicKey({
    mnemonic: ''
  })
);

async function fundL1() {
  const executor = new Wallet(
    config.l1lcd,
    new MnemonicKey({ mnemonic: config.EXECUTOR_MNEMONIC })
  );
  const output = new Wallet(
    config.l1lcd,
    new MnemonicKey({ mnemonic: config.OUTPUT_SUBMITTER_MNEMONIC })
  );
  const batch = new Wallet(
    config.l1lcd,
    new MnemonicKey({ mnemonic: config.BATCH_SUBMITTER_MNEMONIC })
  );
  const challenger = new Wallet(
    config.l1lcd,
    new MnemonicKey({ mnemonic: config.CHALLENGER_MNEMONIC })
  );
  const receiver = new Wallet(
    config.l1lcd,
    new MnemonicKey({
      mnemonic: ''
    })
  );

  const sendMsg = [
    new MsgSend(
      L1_FUNDER.key.accAddress,
      executor.key.accAddress,
      '50000000000uinit'
    ),
    new MsgSend(
      L1_FUNDER.key.accAddress,
      output.key.accAddress,
      '50000000000uinit'
    ),
    new MsgSend(
      L1_FUNDER.key.accAddress,
      batch.key.accAddress,
      '50000000000uinit'
    ),
    new MsgSend(
      L1_FUNDER.key.accAddress,
      challenger.key.accAddress,
      '50000000000uinit'
    ),
    new MsgSend(
      L1_FUNDER.key.accAddress,
      receiver.key.accAddress,
      '100000000000uinit'
    )
  ];
  await sendTx(L1_FUNDER, sendMsg);
}

async function fundL2() {
  const executor = new Wallet(
    config.l2lcd,
    new MnemonicKey({ mnemonic: config.EXECUTOR_MNEMONIC })
  );
  const output = new Wallet(
    config.l2lcd,
    new MnemonicKey({ mnemonic: config.OUTPUT_SUBMITTER_MNEMONIC })
  );
  const batch = new Wallet(
    config.l2lcd,
    new MnemonicKey({ mnemonic: config.BATCH_SUBMITTER_MNEMONIC })
  );
  const challenger = new Wallet(
    config.l2lcd,
    new MnemonicKey({ mnemonic: config.CHALLENGER_MNEMONIC })
  );

  const sendMsg = [
    new MsgSend(L2_FUNDER.key.accAddress, executor.key.accAddress, '1umin'),
    new MsgSend(L2_FUNDER.key.accAddress, output.key.accAddress, '1umin'),
    new MsgSend(L2_FUNDER.key.accAddress, batch.key.accAddress, '1umin'),
    new MsgSend(L2_FUNDER.key.accAddress, challenger.key.accAddress, '1umin')
  ];
  await sendTx(L2_FUNDER, sendMsg);
}

async function startDepositTxBot() {
  const txBot = new TxBot(config.BRIDGE_ID);
  for (;;) {
    const res = await txBot.deposit(
      txBot.l1sender,
      txBot.l2receiver,
      new Coin('uinit', 1_000_000)
    );
    console.log(`Deposited height ${res.height} ${res.txhash}`);
    await delay(1_000);
  }
}

async function main() {
  await startDepositTxBot();
  await fundL1();
  await fundL2();
  console.log('Funded accounts');
}

if (require.main === module) {
  main();
}
