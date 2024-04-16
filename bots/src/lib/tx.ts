import {
  Fee,
  LCDClient,
  Msg,
  WaitTxBroadcastResult,
  Wallet
} from '@initia/initia.js'

export async function sendTx(
  wallet: Wallet,
  msgs: Msg[],
  fee?: Fee,
  accountNumber?: number,
  sequence?: number,
  timeout = 10_000
): Promise<WaitTxBroadcastResult> {
  const signedTx = await wallet.createAndSignTx({
    msgs,
    fee,
    accountNumber,
    sequence
  })
  const broadcastResult = await wallet.lcd.tx.broadcast(signedTx, timeout)
  if (broadcastResult['code']) throw new Error(broadcastResult.raw_log)
  return broadcastResult
}

export async function sendRawTx(
  wallet: Wallet,
  txBytes: string,
  timeout = 10_000
): Promise<WaitTxBroadcastResult> {
  const broadcastResult = await wallet.lcd.tx.broadcast(txBytes, timeout)
  if (broadcastResult['code']) throw new Error(broadcastResult.raw_log)
  return broadcastResult
}

// check whether batch submission interval is met
export async function getLatestBlockHeight(client: LCDClient): Promise<number> {
  const block = await client.tendermint.blockInfo().catch((error) => {
    throw new Error(`Error getting block info from L2: ${error}`)
  })

  return parseInt(block.block.header.height)
}
