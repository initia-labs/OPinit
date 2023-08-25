import config from 'config';
import { Monitor } from './Monitor';
import { CoinInfo, getCoinInfo } from 'lib/lcd';
import {
  AccAddress,
  Coin,
  Msg,
  MsgCreateToken,
  MsgDeposit
} from '@initia/minitia.js';
import { structTagToDenom } from 'lib/util';
import { ExecutorCoinEntity } from 'orm';
import { WalletType, getWallet, TxWallet } from 'lib/wallet';
import { EntityManager } from 'typeorm';
import { RPCSocket } from 'lib/rpc';
import { getDB } from './db';
import winston from 'winston';

export class L1Monitor extends Monitor {
  constructor(public socket: RPCSocket, logger: winston.Logger) {
    super(socket, logger);
    [this.db] = getDB();
  }

  public name(): string {
    return 'executor_l1_monitor';
  }

  public async handleTokenRegisteredEvent(
    manager: EntityManager,
    data: { [key: string]: string }
  ): Promise<[string, CoinInfo]> {
    const coinInfo: CoinInfo = await getCoinInfo(
      data['l1_token'],
      `l2_${data['l2_token']}`
    );
    const l2Denom = coinInfo.denom;
    const coin: ExecutorCoinEntity = {
      l1StructTag: data['l1_token'],
      l1Denom: structTagToDenom(data['l1_token']),
      l2StructTag: `0x1::native_${l2Denom}::Coin`,
      l2Denom
    };

    await this.helper.saveEntity(manager, ExecutorCoinEntity, coin);

    return [l2Denom, coinInfo];
  }

  public async handleEvents(): Promise<void> {
    await this.db.transaction(
      async (transactionalEntityManager: EntityManager) => {
        const msgs: Msg[] = [];
        const wallet: TxWallet = getWallet(WalletType.Executor);
        const events = await this.helper.fetchEvents(
          config.l1lcd,
          this.syncedHeight
        );

        for (const evt of events) {
          const attrMap = this.helper.eventsToAttrMap(evt);
          const data = this.helper.parseData(attrMap);
          if (data['l2_id'] !== config.L2ID) continue;

          switch (attrMap['type_tag']) {
            case '0x1::op_bridge::TokenRegisteredEvent': {
              const [l2Denom, coinInfo] = await this.handleTokenRegisteredEvent(
                transactionalEntityManager,
                data
              );
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
              const denom = `l2_${data['l2_token']}`;
              msgs.push(
                new MsgDeposit(
                  wallet.key.accAddress,
                  AccAddress.fromHex(data['from']),
                  AccAddress.fromHex(data['to']),
                  new Coin(denom, data['amount']),
                  Number.parseInt(data['l1_sequence']),
                  this.syncedHeight + 1,
                  Buffer.from(data['data'])
                )
              );
              break;
            }
          }
        }

        if (msgs.length > 0) {
          await wallet
            .transaction(msgs)
            .then((info) => {
              this.logger.info(
                `Tx submitted in height: ${this.syncedHeight}, txhash: ${info?.txhash}`
              );
            })
            .catch((err) => {
              throw new Error(`Error in L1 Monitor ${err}`);
            });
        }
      }
    );
  }
}
