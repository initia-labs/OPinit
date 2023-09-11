import { BridgeConfig } from './types';
import { getConfig } from 'config';

const config = getConfig();

export interface CoinInfo {
  structTag: string;
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
