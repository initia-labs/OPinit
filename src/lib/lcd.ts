import { BridgeConfig } from './types';
import { getConfig } from 'config';
import { BCS, Wallet, MnemonicKey } from '@initia/initia.js';
import * as crypto from 'crypto';

const config = getConfig();
const bcs = BCS.getInstance();

const executor = new Wallet(
  config.l1lcd,
  new MnemonicKey({ mnemonic: config.EXECUTOR_MNEMONIC })
);

export interface FAMetadata {
  name: string;
  symbol: string;
  decimals: string;
  icon_url: string;
  reference_url: string;
}

export interface CoinInfo {
  denom: string;
  name: string;
  symbol: string;
  decimals: number;
}

export async function fetchBridgeConfig(): Promise<BridgeConfig> {
  const cfg = await config.l1lcd.move.viewFunction<BridgeConfig>(
    '0x1',
    'op_output',
    'get_config_store',
    [],
    [
      bcs.serialize('address', executor.key.accAddress),
      bcs.serialize('string', config.L2ID)
    ]
  );
  return cfg;
}

const OBJECT_DERIVED_SCHEME = 0xfc;
const OBJECT_FROM_SEED_ADDRESS_SCHEME = 0xfe;
const BRIDGE_PREFIX = 0xf2;

export function normalizeMetadata(addr: string) {
  return addr.startsWith('0x') ? addr : '0x' + addr;
}

export function computeBridgeMetadata(creator:string, l2Id: string) {
  const addrBytes = Buffer.from(
    bcs.serialize('address', creator),
    'base64'
  ).toJSON().data;
  const combinedSeed = [BRIDGE_PREFIX, ...Buffer.from(l2Id)];
  const combinedBytes = [
    ...addrBytes,
    ...combinedSeed,
    OBJECT_FROM_SEED_ADDRESS_SCHEME
  ];

  const hash = crypto
    .createHash('SHA3-256')
    .update(Buffer.from(combinedBytes))
    .digest();
  return normalizeMetadata(hash.toString('hex'));
}

export function computePrimaryMetadata(owner: string, coinMetadata: string) {
  const addrBytes = Buffer.from(
    bcs.serialize('address', owner),
    'base64'
  ).toJSON().data;
  const seed = Buffer.from(coinMetadata, 'ascii').toJSON().data;
  const combinedBytes = [...addrBytes, ...seed, OBJECT_DERIVED_SCHEME];

  const hash = crypto
    .createHash('SHA3-256')
    .update(Buffer.from(combinedBytes))
    .digest();
  return hash.toString('hex');
}

export function computeCoinMetadata(creator: string, symbol: string): string {
  const addrBytes = Buffer.from(
    bcs.serialize('address', creator),
    'base64'
  ).toJSON().data;
  const seed = Buffer.from(symbol, 'ascii').toJSON().data;
  const combinedBytes = [
    ...addrBytes,
    ...seed,
    OBJECT_FROM_SEED_ADDRESS_SCHEME
  ];

  const hash = crypto
    .createHash('SHA3-256')
    .update(Buffer.from(combinedBytes))
    .digest();
  return hash.toString('hex');
}

export async function resolveFAMetadata(
  lcd: any,
  metadata: string
): Promise<CoinInfo> {
  const resourceData = await lcd.move.resource(
    metadata,
    '0x1::fungible_asset::Metadata'
  );
  const symbol = resourceData.data.symbol;
  const sanitizedMetadata = metadata.startsWith('0x')
    ? metadata.slice(2)
    : metadata;
  const isNative = sanitizedMetadata === computeCoinMetadata('0x1', symbol);
  const denom = isNative ? symbol : `move/${sanitizedMetadata}`;

  return {
    name: resourceData.data.name,
    symbol: symbol,
    denom: denom,
    decimals: Number.parseInt(resourceData.data.decimals, 10)
  };
}
