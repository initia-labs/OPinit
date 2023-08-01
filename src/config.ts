import { LCDClient as MinitiaLCDClient } from '@initia/minitia.js'
import { LCDClient as InitiaLCDClient } from '@initia/initia.js'
const defaultConfig = {
    EXECUTOR_PORT: '3000',
    BATCH_PORT: '3001',
    L1_LCD_URI: 'https://stone-rest.initia.tech',
    L1_RPC_URI: 'https://stone-rpc.initia.tech',
    L2_LCD_URI: 'https://minitia-rest.initia.tech',
    L2_RPC_URI: 'https://minitia-rpc.initia.tech',
    L2ID: '',
    OUTPUT_SUBMITTER_MNEMONIC: '',
    EXECUTOR_MNEMONIC: '',
    BATCH_SUBMITTER_MNEMONIC: '',
    CHALLENGER_MNEMONIC: '',
    USE_LOG_FILE: 'false',
}

const {
    EXECUTOR_PORT,
    BATCH_PORT,
    L1_LCD_URI,
    L1_RPC_URI,
    L2_LCD_URI,
    L2_RPC_URI,
    L2ID,
    OUTPUT_SUBMITTER_MNEMONIC,
    EXECUTOR_MNEMONIC,
    BATCH_SUBMITTER_MNEMONIC,
    CHALLENGER_MNEMONIC,
    USE_LOG_FILE,
} = {...defaultConfig, ...process.env}

const config = {
    EXECUTOR_PORT: parseInt(EXECUTOR_PORT),
    BATCH_PORT: parseInt(BATCH_PORT),
    L1_LCD_URI,
    L1_RPC_URI,
    L2_LCD_URI,
    L2_RPC_URI,
    L2ID,
    OUTPUT_SUBMITTER_MNEMONIC,
    EXECUTOR_MNEMONIC,
    BATCH_SUBMITTER_MNEMONIC,
    CHALLENGER_MNEMONIC,
    l1lcd: new InitiaLCDClient(L1_LCD_URI),
    l2lcd: new MinitiaLCDClient(L2_LCD_URI, {gasPrices: '0umin', gasAdjustment: "1.75"}),
    USE_LOG_FILE: !!JSON.parse(USE_LOG_FILE),
    EXCLUDED_ROUTES: [],
}

export default config