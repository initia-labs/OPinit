import { LCDClient } from '@initia/initia.js';
import * as dotenv from 'dotenv';

const envFile =
  ( process.env.NODE_ENV === 'test' || !process.env.WORKER_NAME ) ? `.env` : `.env.${process.env.WORKER_NAME}`;

console.log('activate ', envFile);
dotenv.config({ path: envFile });

const {
  EXECUTOR_PORT,
  BATCH_PORT,
  L1_LCD_URI,
  L1_RPC_URI,
  L2_LCD_URI,
  L2_RPC_URI,
  EXECUTOR_URI,
  BRIDGE_ID,
  OUTPUT_SUBMITTER_MNEMONIC,
  EXECUTOR_MNEMONIC,
  BATCH_SUBMITTER_MNEMONIC,
  CHALLENGER_MNEMONIC,
  USE_LOG_FILE,
  L2_GAS_PRICES,
  L1_CHAIN_ID,
  L2_CHAIN_ID,
  SLACK_WEB_HOOK,
  SUBMISSION_INTERVAL,
  FINALIZATION_PERIOD,
  IBC_METADATA,
  DELETE_OUTPUT_PROPOSAL
} = process.env;

export const config = {
  EXECUTOR_PORT: EXECUTOR_PORT ? parseInt(EXECUTOR_PORT) : 5000,
  BATCH_PORT: BATCH_PORT ? parseInt(BATCH_PORT) : 5001,
  L1_LCD_URI: L1_LCD_URI ? L1_LCD_URI.split(',') : ['http://localhost:1317'],
  L1_RPC_URI: L1_RPC_URI ? L1_RPC_URI.split(',') : ['http://localhost:26657'],
  L2_LCD_URI: L2_LCD_URI ? L2_LCD_URI.split(',') : ['http://localhost:1317'],
  L2_RPC_URI: L2_RPC_URI ? L2_RPC_URI.split(',') : ['http://localhost:26657'],
  EXECUTOR_URI: EXECUTOR_URI || 'http://localhost:5000',
  BRIDGE_ID: BRIDGE_ID ? parseInt(BRIDGE_ID) : 1,
  OUTPUT_SUBMITTER_MNEMONIC: OUTPUT_SUBMITTER_MNEMONIC
    ? OUTPUT_SUBMITTER_MNEMONIC.replace(/'/g, '')
    : '',
  EXECUTOR_MNEMONIC: EXECUTOR_MNEMONIC
    ? EXECUTOR_MNEMONIC.replace(/'/g, '')
    : '',
  BATCH_SUBMITTER_MNEMONIC: BATCH_SUBMITTER_MNEMONIC
    ? BATCH_SUBMITTER_MNEMONIC.replace(/'/g, '')
    : '',
  CHALLENGER_MNEMONIC: CHALLENGER_MNEMONIC
    ? CHALLENGER_MNEMONIC.replace(/'/g, '')
    : '',
  USE_LOG_FILE: USE_LOG_FILE ? JSON.parse(USE_LOG_FILE) : false,
  L1_CHAIN_ID: L1_CHAIN_ID ? L1_CHAIN_ID : 'local-initia',
  L2_CHAIN_ID: L2_CHAIN_ID ? L2_CHAIN_ID : 'local-minitia',
  l1lcd: new LCDClient(
    L1_LCD_URI ? L1_LCD_URI.split(',')[0] : 'http://localhost:1317',
    {
      gasPrices: '0.15uinit',
      gasAdjustment: '2'
    }
  ),
  l2lcd: new LCDClient(
    L2_LCD_URI ? L2_LCD_URI.split(',')[0] : 'http://localhost:1317',
    {
      gasPrices: L2_GAS_PRICES || '0.15umin',
      gasAdjustment: '2'
    }
  ),
  SLACK_WEB_HOOK: SLACK_WEB_HOOK ? SLACK_WEB_HOOK : '',
  SUBMISSION_INTERVAL: SUBMISSION_INTERVAL
    ? parseInt(SUBMISSION_INTERVAL)
    : 3600,
  FINALIZATION_PERIOD: FINALIZATION_PERIOD
    ? parseInt(FINALIZATION_PERIOD)
    : 3600,
  IBC_METADATA: IBC_METADATA ? IBC_METADATA : '',
  DELETE_OUTPUT_PROPOSAL: DELETE_OUTPUT_PROPOSAL
    ? DELETE_OUTPUT_PROPOSAL
    : 'false'
};

export const INTERVAL_BATCH = 10_000;
export const INTERVAL_MONITOR = 100;
export const INTERVAL_OUTPUT = 10_000;
