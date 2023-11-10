import { LCDClient, TxSearchResult, TxInfo } from '@initia/initia.js';

export async function txSearch(
  lcd: LCDClient,
  height: number
): Promise<TxInfo[]> {
  let page = 0;
  let totalPage = 0;
  const limit = 100;
  const txInfos: TxInfo[] = [];

  do {
    const txResult: TxSearchResult = await lcd.tx.search({
      events: [
        {
          key: 'tx.height',
          value: height.toFixed()
        }
      ],
      'pagination.limit': limit.toFixed(),
      'pagination.offset': (page * limit).toFixed()
    });
    txInfos.push(...txResult.txs);
    if (txResult.pagination) totalPage = txResult.pagination.total / limit;
  } while (++page < totalPage);

  return txInfos;
}
