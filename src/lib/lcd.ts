import { delay } from 'bluebird';
import { TxInfo, Wallet, Msg } from '@initia/minitia.js';
import { BridgeConfig } from './types';
import config from '../config';

export async function transaction(
  wallet: Wallet,
  msgs: Msg[],
  accountNumber?: number,
  sequence?: number
): Promise<TxInfo | undefined> {
  const signedTx = await wallet.createAndSignTx({
    msgs,
    accountNumber,
    sequence
  });
  const broadcastResult = await wallet.lcd.tx.broadcast(signedTx);
  if (broadcastResult['code']) throw new Error(broadcastResult.raw_log);
  return checkTx(wallet, broadcastResult.txhash);
}

export async function checkTx(
  wallet: Wallet,
  txHash: string,
  timeout = 60000
): Promise<TxInfo | undefined> {
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeout) {
    const txInfo = await wallet.lcd.tx.txInfo(txHash);
    if (txInfo) return txInfo;
    await delay(1000);
  }
}

export async function fetchBridgeConfig(): Promise<BridgeConfig> {
  const cfg = await config.l1lcd.move.viewFunction<BridgeConfig>(
    '0x1',
    'op_output',
    'get_config_store',
    [config.L2ID],
    []
  );
  return cfg;
}

export async function getCoinInfo(
  structTag: string,
  l2Denom: string
): Promise<CoinInfo> {
  const address = structTag.split('::')[0];
  const resource = await config.l1lcd.move.resource<{
    name: string;
    symbol: string;
    decimals: number;
  }>(address, `0x1::coin::CoinInfo<${structTag}>`);

  return {
    structTag,
    denom: l2Denom,
    name: resource.data.name,
    symbol: resource.data.symbol,
    decimals: resource.data.decimals
  };
}

export interface CoinInfo {
  structTag: string;
  denom: string;
  name: string;
  symbol: string;
  decimals: number;
}
