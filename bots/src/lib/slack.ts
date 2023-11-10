import { Wallet } from '@initia/initia.js';
import axios from 'axios';
import BigNumber from 'bignumber.js';
import { getConfig } from 'config';
import * as http from 'http';
import * as https from 'https';
import FailedTxEntity from 'orm/executor/FailedTxEntity';

const config = getConfig();

const { SLACK_WEB_HOOK } = process.env;

export function buildNotEnoughBalanceNotification(
  wallet: Wallet,
  balance: number,
  denom: string
): { text: string } {
  let notification = '```';
  notification += `[WARN] Enough Balance Notification\n`;
  notification += `\n`;
  notification += `Endpoint: ${wallet.lcd.URL}\n`;
  notification += `Address : ${wallet.key.accAddress}\n`;
  notification += `Balance : ${new BigNumber(balance)
    .div(1e6)
    .toFixed(6)} ${denom}\n`;
  notification += '```';
  const text = `${notification}`;
  return {
    text
  };
}

export function buildFailedTxNotification(data: FailedTxEntity): {
  text: string;
} {
  let notification = '```';
  notification += `[WARN] Bridge Processed Tx Notification\n`;

  notification += `[L1] ${config.L1_CHAIN_ID} => [L2] ${config.L2_CHAIN_ID}\n`;
  notification += `\n`;
  notification += `Bridge ID: ${data.bridgeId}\n`;
  notification += `Sequence:  ${data.sequence}\n`;
  notification += `Sender:    ${data.sender}\n`;
  notification += `To:        ${data.receiver}\n`;
  notification += `\n`;
  notification += `Amount:    ${new BigNumber(data.amount)
    .div(1e6)
    .toFixed(6)} ${data.l1Denom}\n`;
  notification += `\n`;
  notification += `L1 Height: ${data.l1Height}\n`;
  notification += `Error    : ${data.error}\n`;
  notification += '```';
  const text = `${notification}`;

  return {
    text
  };
}

const ax = axios.create({
  httpAgent: new http.Agent({ keepAlive: true }),
  httpsAgent: new https.Agent({ keepAlive: true }),
  timeout: 15000
});

export async function notifySlack(text: { text: string }) {
  if (SLACK_WEB_HOOK == undefined || SLACK_WEB_HOOK == '') return;
  await ax.post(SLACK_WEB_HOOK, text).catch(() => {
    console.error('Slack Notification Error');
  });
}
