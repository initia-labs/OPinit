import { LCDClient as MinitiaLCDClient } from '@initia/minitia.js';
import { LCDClient as InitiaLCDClient } from '@initia/initia.js';

interface ConfigInterface {
  EXECUTOR_PORT: number;
  BATCH_PORT: number;
  L1_LCD_URI: string;
  L1_RPC_URI: string;
  L2_LCD_URI: string;
  L2_RPC_URI: string;
  EXECUTOR_URI: string;
  L2ID: string;
  OUTPUT_SUBMITTER_MNEMONIC: string;
  EXECUTOR_MNEMONIC: string;
  BATCH_SUBMITTER_MNEMONIC: string;
  CHALLENGER_MNEMONIC: string;
  USE_LOG_FILE: boolean;
  l1lcd: InitiaLCDClient;
  l2lcd: MinitiaLCDClient;
  EXCLUDED_ROUTES: string[];
}

const defaultConfig = {
  EXECUTOR_PORT: '3000',
  BATCH_PORT: '3001',
  L1_LCD_URI: 'https://stone-rest.initia.tech',
  L1_RPC_URI: 'https://stone-rpc.initia.tech',
  L2_LCD_URI: 'https://minitia-rest.initia.tech',
  L2_RPC_URI: 'https://minitia-rpc.initia.tech',
  EXECUTOR_URI: 'https://minitia-executor.initia.tech',
  L2ID: '',
  OUTPUT_SUBMITTER_MNEMONIC: '',
  EXECUTOR_MNEMONIC: '',
  BATCH_SUBMITTER_MNEMONIC: '',
  CHALLENGER_MNEMONIC: '',
  USE_LOG_FILE: 'false'
};

export class Config implements ConfigInterface {
  private static instance: Config;

  EXECUTOR_PORT: number;
  BATCH_PORT: number;
  L1_LCD_URI: string;
  L1_RPC_URI: string;
  L2_LCD_URI: string;
  L2_RPC_URI: string;
  EXECUTOR_URI: string;
  L2ID: string;
  OUTPUT_SUBMITTER_MNEMONIC: string;
  EXECUTOR_MNEMONIC: string;
  BATCH_SUBMITTER_MNEMONIC: string;
  CHALLENGER_MNEMONIC: string;
  USE_LOG_FILE: boolean;
  l1lcd: InitiaLCDClient;
  l2lcd: MinitiaLCDClient;
  EXCLUDED_ROUTES: string[] = [];

  private constructor() {
    const {
      EXECUTOR_PORT,
      BATCH_PORT,
      L1_LCD_URI,
      L1_RPC_URI,
      L2_LCD_URI,
      L2_RPC_URI,
      EXECUTOR_URI,
      L2ID,
      OUTPUT_SUBMITTER_MNEMONIC,
      EXECUTOR_MNEMONIC,
      BATCH_SUBMITTER_MNEMONIC,
      CHALLENGER_MNEMONIC,
      USE_LOG_FILE
    } = { ...defaultConfig, ...process.env };

    this.EXECUTOR_PORT = parseInt(EXECUTOR_PORT);
    this.BATCH_PORT = parseInt(BATCH_PORT);
    this.L1_LCD_URI = L1_LCD_URI;
    this.L1_RPC_URI = L1_RPC_URI;
    this.L2_LCD_URI = L2_LCD_URI;
    this.L2_RPC_URI = L2_RPC_URI;
    this.EXECUTOR_URI = EXECUTOR_URI;
    this.L2ID = L2ID;
    this.OUTPUT_SUBMITTER_MNEMONIC = OUTPUT_SUBMITTER_MNEMONIC;
    this.EXECUTOR_MNEMONIC = EXECUTOR_MNEMONIC;
    this.BATCH_SUBMITTER_MNEMONIC = BATCH_SUBMITTER_MNEMONIC;
    this.CHALLENGER_MNEMONIC = CHALLENGER_MNEMONIC;
    this.USE_LOG_FILE = !!JSON.parse(USE_LOG_FILE);
    this.l1lcd = new InitiaLCDClient(L1_LCD_URI);
    this.l2lcd = new MinitiaLCDClient(L2_LCD_URI, {
      gasPrices: '0umin',
      gasAdjustment: '1.75'
    });
  }

  public static getConfig(): Config {
    if (!Config.instance) {
      Config.instance = new Config();
    }
    return Config.instance;
  }

  public static updateConfig(newConfig: Partial<ConfigInterface>) {
    Config.instance = { ...Config.instance, ...newConfig };
  }
}

export function getConfig() {
  if (process.env.DEVELOPMENT_MODE === 'test') {
    process.env.TYPEORM_HOST = 'localhost';
    process.env.TYPEORM_USERNAME = 'user';
    process.env.TYPEORM_PASSWORD = 'password';
    process.env.TYPEORM_DATABASE = 'rollup';
    process.env.TYPEORM_PORT = '5433';

    const testConfig = {
      EXECUTOR_PORT: 3000,
      BATCH_PORT: 3001,
      L1_LCD_URI: 'http://localhost:1317',
      L1_RPC_URI: 'http://localhost:26657',
      L2_LCD_URI: 'http://localhost:1318',
      L2_RPC_URI: 'http://localhost:26658',
      EXECUTOR_URI: 'http://localhost:3000',
      L2ID: '0x56ccf33c45b99546cd1da172cf6849395bbf8573::s10ta1::Minitia',
      TYPEORM_HOST: 'http://localhost:5433'
    };
    Config.updateConfig({
      ...testConfig,
      l1lcd: new InitiaLCDClient(testConfig.L1_LCD_URI, {
        gasAdjustment: '1.75'
      }),
      l2lcd: new MinitiaLCDClient(testConfig.L2_LCD_URI, {
        gasPrices: '0.15umin',
        gasAdjustment: '1.75'
      })
    });
  }

  return Config.getConfig();
}

const config = Config.getConfig();
export default config;

export const INTERVAL_BATCH = 10000;
export const INTERVAL_MONITOR = 100;
export const INTERVAL_OUTPUT = 10000;
