import { logger } from 'lib/logger';
import config from 'config';
import { Monitor } from './Monitor';
import { getCoinInfo } from 'lib/lcd';
import {
  AccAddress,
  Coin,
  Msg,
  MsgCreateToken,
  MsgDeposit
} from '@initia/minitia.js';
import { getL2Denom, structTagToDenom } from 'lib/util';
import { CoinEntity } from 'orm';
import { WalletType, getWallet, TxWallet } from 'lib/wallet';
import { EntityManager } from 'typeorm';

export class L1Monitor extends Monitor {
  public name(): string {
    return 'l1_monitor';
  }

  public color(): string {
    return 'blue';
  }

  public async handleEvents(): Promise<void> {
    await this.db.transaction(
      async (transactionalEntityManager: EntityManager) => {
        const msgs: Msg[] = [];
        const wallet: TxWallet = getWallet(WalletType.Executor);
        const searchRes = await config.l1lcd.tx.search({
          events: [
            { key: 'tx.height', value: (this.syncedHeight + 1).toString() }
          ]
        });
        const events = searchRes.txs
          .flatMap((tx) => tx.logs ?? [])
          .flatMap((log) => log.events);
        for (const evt of events) {
          if (evt.type !== 'move') continue;

          const attrMap: { [key: string]: string } = evt.attributes.reduce(
            (obj, attr) => {
              obj[attr.key] = attr.value;
              return obj;
            },
            {}
          );

          const data: { [key: string]: string } = JSON.parse(attrMap['data']);
          if (data['l2_id'] !== config.L2ID) continue;

          switch (attrMap['type_tag']) {
            case '0x1::op_bridge::TokenRegisteredEvent': {
              // handle token registered event
              const coinInfo = await getCoinInfo(
                data['l1_token'],
                Buffer.from(data['l2_token'])
              );
              const l2Denom = coinInfo.denom;
              const coin: CoinEntity = {
                l1StructTag: data['l1_token'],
                l1Denom: structTagToDenom(data['l1_token']),
                l2StructTag: `0x1::native_${l2Denom}::Coin`,
                l2Denom
              };

              await transactionalEntityManager
                .getRepository(CoinEntity)
                .save(coin);
              msgs.push(
                new MsgCreateToken(
                  wallet.key.accAddress,
                  l2Denom,
                  coinInfo.name,
                  coinInfo.symbol,
                  coinInfo.decimals
                )
              );
              break;
            }
            case '0x1::op_bridge::TokenBridgeInitiatedEvent': {
              // handle token bridge initiated event
              const denom = getL2Denom(Buffer.from(data['l2_token']));
              msgs.push(
                new MsgDeposit(
                  wallet.key.accAddress,
                  AccAddress.fromHex(data['from']),
                  AccAddress.fromHex(data['to']),
                  new Coin(denom, data['amount']),
                  Number.parseInt(data['l1_sequence'])
                )
              );
              break;
            }
          }
        }

        if (msgs.length > 0) {
          await wallet
            .transaction(msgs)
            .then((info) => logger.info(`Tx submitted: ${info?.txhash}`))
            .catch((err) => logger.error(`Err in L1 Monitor ${err}`));
        }
      }
    );
  }
}
