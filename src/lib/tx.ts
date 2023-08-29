import { delay } from 'bluebird';
import { Wallet, LCDClient, TxInfo, Msg } from '@initia/minitia.js';

export async function transaction(
  wallet: Wallet,
  msgs: Msg[],
  accountNumber?: number,
  sequence?: number
): Promise<TxInfo | undefined> {
  try {
    const signedTx = await wallet.createAndSignTx({
      msgs,
      accountNumber,
      sequence
    });
    const broadcastResult = await wallet.lcd.tx.broadcast(signedTx);
    if (broadcastResult['code']) throw new Error(broadcastResult.raw_log);
    return checkTx(wallet.lcd, broadcastResult.txhash);
  } catch (err) {
    throw new Error(`Failed to execute transaction: ${err.message}`);
  }
}

export async function checkTx(
  lcd: LCDClient,
  txHash: string,
  timeout = 60000
): Promise<TxInfo | undefined> {
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeout) {
    try {
      const txInfo = await lcd.tx.txInfo(txHash);
      if (txInfo) return txInfo;
      await delay(1000);
    } catch (err) {
      throw new Error(`Failed to check transaction status: ${err.message}`);
    }
  }

  throw new Error('Transaction checking timed out');
}

// check whether batch submission interval is met
export async function getLatestBlockHeight(client: LCDClient): Promise<number> {
  const block = await client.tendermint.blockInfo().catch((error) => {
    throw new Error(`Error getting block info from L2: ${error}`);
  });

  return parseInt(block.block.header.height);
}
