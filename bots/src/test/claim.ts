import { config } from 'config';
import { TxBot } from './utils/TxBot';
import { Coin } from '@initia/initia.js';

const txBot = new TxBot(config.BRIDGE_ID);

async function main() {
  try {
    // await withdraw()
    await claim(3, 21); // set sequence, outputIndex
  } catch (err) {
    console.log(err);
  }
}

export async function claim(sequence: number, outputIndex: number) {
  const beforeBalance = await config.l1lcd.bank.balance(
    txBot.l1sender.key.accAddress
  );

  const res = await txBot.claim(txBot.l1sender, sequence, outputIndex);

  const afterBalance = await config.l1lcd.bank.balance(
    txBot.l1sender.key.accAddress
  );

  console.log(
    `claimed : ${afterBalance[0]
      .get('uinit')
      ?.sub(beforeBalance[0].get('uinit') ?? 0)} in hash ${res.txhash}`
  );
}

export async function withdraw() {
  const pair = await config.l1lcd.ophost.tokenPairByL1Denom(
    config.BRIDGE_ID,
    'uinit'
  );

  const beforeBalance = await config.l2lcd.bank.balance(
    txBot.l2sender.key.accAddress
  );

  const res = await txBot.withdrawal(
    txBot.l2sender,
    txBot.l1sender,
    new Coin(pair.l2_denom, 1_000_000)
  );

  const afterBalance = await config.l2lcd.bank.balance(
    txBot.l2sender.key.accAddress
  );

  console.log(
    `withdraw: ${beforeBalance[0]
      .get(pair.l2_denom)
      ?.sub(afterBalance[0].get(pair.l2_denom) ?? 0)} in hash ${res.txhash}`
  );
}

if (require.main === module) {
  main();
}
