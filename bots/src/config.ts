import { LCDClient } from '@initia/initia.js';

interface ConfigInterface {
  EXECUTOR_PORT: number;
  BATCH_PORT: number;
  L1_LCD_URI: string[];
  L1_RPC_URI: string[];
  L2_LCD_URI: string[];
  L2_RPC_URI: string[];
  EXECUTOR_URI: string; // only for test
  BRIDGE_ID: number;
  OUTPUT_SUBMITTER_MNEMONIC: string;
  EXECUTOR_MNEMONIC: string;
  BATCH_SUBMITTER_MNEMONIC: string;
  CHALLENGER_MNEMONIC: string;
  USE_LOG_FILE: boolean;
  l1lcd: LCDClient;
  l2lcd: LCDClient;
  L2_DENOM: string;
}

const defaultConfig = {
  EXECUTOR_PORT: '5000',
  BATCH_PORT: '5001',
  L1_LCD_URI: 'https://stone-rest.initia.tech',
  L1_RPC_URI: 'https://stone-rpc.initia.tech',
  L2_LCD_URI: 'https://minitia-rest.initia.tech',
  L2_RPC_URI: 'https://minitia-rpc.initia.tech',
  EXECUTOR_URI: 'https://minitia-executor.initia.tech',
  BRIDGE_ID: '',
  OUTPUT_SUBMITTER_MNEMONIC: '',
  EXECUTOR_MNEMONIC: '',
  BATCH_SUBMITTER_MNEMONIC: '',
  CHALLENGER_MNEMONIC: '',
  USE_LOG_FILE: 'false',
  L2_DENOM: 'umin',
  L1_CHAIN_ID: '',
  L2_CHAIN_ID: ''
};

export class Config implements ConfigInterface {
  private static instance: Config;

  EXECUTOR_PORT: number;
  BATCH_PORT: number;
  L1_LCD_URI: string[];
  L1_RPC_URI: string[];
  L2_LCD_URI: string[];
  L2_RPC_URI: string[];
  EXECUTOR_URI: string;
  BRIDGE_ID: number;
  OUTPUT_SUBMITTER_MNEMONIC: string;
  EXECUTOR_MNEMONIC: string;
  BATCH_SUBMITTER_MNEMONIC: string;
  CHALLENGER_MNEMONIC: string;
  USE_LOG_FILE: boolean;
  l1lcd: LCDClient;
  l2lcd: LCDClient;
  L2_DENOM: string;
  L1_CHAIN_ID: string;
  L2_CHAIN_ID: string;

  private constructor() {
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
      L2_DENOM,
      L1_CHAIN_ID,
      L2_CHAIN_ID
    } = { ...defaultConfig, ...process.env };

    this.EXECUTOR_PORT = parseInt(EXECUTOR_PORT);
    this.BATCH_PORT = parseInt(BATCH_PORT);
    this.L1_LCD_URI = L1_LCD_URI.split(',');
    this.L1_RPC_URI = L1_RPC_URI.split(',');
    this.L2_LCD_URI = L2_LCD_URI.split(',');
    this.L2_RPC_URI = L2_RPC_URI.split(',');
    this.EXECUTOR_URI = EXECUTOR_URI;
    this.BRIDGE_ID = parseInt(BRIDGE_ID);
    this.OUTPUT_SUBMITTER_MNEMONIC = OUTPUT_SUBMITTER_MNEMONIC.replace(
      /'/g,
      ''
    );
    this.EXECUTOR_MNEMONIC = EXECUTOR_MNEMONIC.replace(/'/g, '');
    this.BATCH_SUBMITTER_MNEMONIC = BATCH_SUBMITTER_MNEMONIC.replace(/'/g, '');
    this.CHALLENGER_MNEMONIC = CHALLENGER_MNEMONIC.replace(/'/g, '');
    this.USE_LOG_FILE = !!JSON.parse(USE_LOG_FILE);
    this.l1lcd = new LCDClient(this.L1_LCD_URI[0], {
      gasPrices: '0.15uinit',
      gasAdjustment: '2'
    });

    this.L2_DENOM = L2_DENOM;
    this.L1_CHAIN_ID = L1_CHAIN_ID;
    this.L2_CHAIN_ID = L2_CHAIN_ID;

    this.l2lcd = new LCDClient(this.L2_LCD_URI[0], {
      gasPrices: `0.15${this.L2_DENOM}`,
      gasAdjustment: '2'
    });
  }

  public static getConfig(): Config {
    if (!Config.instance) {
      Config.instance = new Config();
    }
    return Config.instance;
  }
}

export function getConfig() {
  return Config.getConfig();
}

const config = Config.getConfig();
export default config;

export const INTERVAL_BATCH = 10_000;
export const INTERVAL_MONITOR = 100;
export const INTERVAL_OUTPUT = 10_000;
