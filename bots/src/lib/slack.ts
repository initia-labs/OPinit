import { Wallet } from '@initia/initia.js'
import axios from 'axios'
import BigNumber from 'bignumber.js'
import { config } from '../config'
import * as http from 'http'
import * as https from 'https'
import UnconfirmedTxEntity from '../orm/executor/UnconfirmedTxEntity'
import { ChallengedOutputEntity } from '../orm/index'

const postedKeys = new Set<string>()

const ax = axios.create({
  httpAgent: new http.Agent({ keepAlive: true }),
  httpsAgent: new https.Agent({ keepAlive: true }),
  timeout: 15000
})

export async function notifySlack(key: string, text: { text: string }, isError: boolean = true) {
  if (config.SLACK_WEB_HOOK === undefined || config.SLACK_WEB_HOOK === '') return

  const keyExists = postedKeys.has(key)

  if (isError) {
      if (!keyExists) {
        await ax.post(config.SLACK_WEB_HOOK, text)
        postedKeys.add(key)
      }
  } else {
      if (keyExists) { 
        await ax.post(config.SLACK_WEB_HOOK, text)
        postedKeys.delete(key)
      }
  }
}

export function buildResolveErrorNotification(description: string): { text: string } {
  let notification = '```'
  notification += `[INFO] Error Resolved Notification\n`
  notification += `\n`
  notification += `${description}\n`
  notification += '```'
  return {
    text: notification
  }
}


export function buildNotEnoughBalanceNotification(
  wallet: Wallet,
  balance: number,
  denom: string
): { text: string } {
  let notification = '```'
  notification += `[WARN] Enough Balance Notification\n`
  notification += `\n`
  notification += `Chain ID: ${wallet.lcd.config.chainId}\n`
  notification += `Endpoint: ${wallet.lcd.URL}\n`
  notification += `Address : ${wallet.key.accAddress}\n`
  notification += `Balance : ${new BigNumber(balance)
    .div(1e6)
    .toFixed(6)} ${denom}\n`
  notification += '```'
  const text = `${notification}`
  return {
    text
  }
}

export function buildFailedTxNotification(data: UnconfirmedTxEntity): {
  text: string;
} {
  let notification = '```'
  notification += `[WARN] Bridge Processed Tx Notification\n`

  notification += `[L1] ${config.L1_CHAIN_ID} => [L2] ${config.L2_CHAIN_ID}\n`
  notification += `\n`
  notification += `Bridge ID: ${data.bridgeId}\n`
  notification += `Sequence:  ${data.sequence}\n`
  notification += `Sender:    ${data.sender}\n`
  notification += `To:        ${data.receiver}\n`
  notification += `\n`
  notification += `Amount:    ${new BigNumber(data.amount)
    .div(1e6)
    .toFixed(6)} ${data.l1Denom}\n`
  notification += `\n`
  notification += `L1 Height: ${data.l1Height}\n`
  notification += `Error    : ${data.error}\n`
  notification += '```'
  const text = `${notification}`

  return {
    text
  }
}

export function buildChallengerNotification(
  challengedOutput: ChallengedOutputEntity
): { text: string } {
  let notification = '```'
  notification += `[WARN] Challenger Notification\n`
  notification += `\n`
  notification += `Bridge ID   : ${challengedOutput.bridgeId}\n`
  notification += `OutputIndex : ${challengedOutput.outputIndex}\n`
  notification += `Reason      : ${challengedOutput.reason}\n`
  notification += '```'
  const text = `${notification}`

  return {
    text
  }
}
