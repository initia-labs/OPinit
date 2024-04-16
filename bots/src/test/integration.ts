import Bridge from './utils/Bridge'
import { config } from '../config'
import { TxBot } from './utils/TxBot'
import { Coin } from '@initia/initia.js'
import { startBatch } from '../worker/batchSubmitter'
import { startExecutor } from '../worker/bridgeExecutor'
import { startOutput } from '../worker/outputSubmitter'
import { delay } from 'bluebird'
import { getTokenPairByL1Denom } from '../lib/query'

const SUBMISSION_INTERVAL = 5
const FINALIZATION_PERIOD = 5
const DEPOSIT_AMOUNT = 1_000_000
const DEPOSIT_INTERVAL_MS = 100

async function setupBridge(submissionInterval: number, finalizedTime: number) {
  const bridge = new Bridge(submissionInterval, finalizedTime)
  const relayerMetadata = ''
  await bridge.clearDB()
  await bridge.tx(relayerMetadata)
  console.log('Bridge deployed')
}

async function startBot() {
  try {
    await Promise.all([startBatch(), startExecutor(), startOutput()])
  } catch (err) {
    console.log(err)
  }
}

async function startDepositTxBot() {
  const txBot = new TxBot(config.BRIDGE_ID)
  const pair = await getTokenPairByL1Denom('uinit')
  for (;;) {
    const balance = await config.l2lcd.bank.balanceByDenom(
      txBot.l2receiver.key.accAddress,
      pair.l2_denom
    )
    const res = await txBot.deposit(
      txBot.l1sender,
      txBot.l2receiver,
      new Coin('uinit', DEPOSIT_AMOUNT)
    )
    console.log(
      `[DepositBot] Deposited height ${res.height} to ${txBot.l2receiver.key.accAddress} ${balance?.amount}`
    )
    await delay(DEPOSIT_INTERVAL_MS)
  }
}

async function main() {
  try {
    await setupBridge(SUBMISSION_INTERVAL, FINALIZATION_PERIOD)
    await startBot()
    await startDepositTxBot()
  } catch (err) {
    console.log(err)
  }
}

if (require.main === module) {
  main()
}
