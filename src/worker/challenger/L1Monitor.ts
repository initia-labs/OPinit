import config from 'config';
import { Monitor } from './Monitor';
import { structTagToDenom } from 'lib/util';
import { ChallengerCoinEntity, DepositTxEntity } from 'orm';
import { getCoinInfo } from 'lib/lcd';
import { EntityManager } from 'typeorm';

export class L1Monitor extends Monitor {
  public name(): string {
    return 'challenger_l1_monitor';
  }

  public color(): string {
    return 'blue';
  }

  public async handleEvents(): Promise<void> {
    await this.db.transaction(
      async (transactionalEntityManager: EntityManager) => {
        const searchRes = await config.l1lcd.tx.search({
          events: [
            { key: 'tx.height', value: (this.syncedHeight + 1).toString() }
          ]
        });
        const events = searchRes.txs
          .flatMap((tx) => tx.logs ?? [])
          .flatMap((log) => log.events);

        const lastIndex = await this.getLastOutputIndex(
          transactionalEntityManager
        );

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
                `l2_${data['l2_token']}`
              );
              const l2Denom = coinInfo.denom;
              const coin: ChallengerCoinEntity = {
                l1StructTag: data['l1_token'],
                l1Denom: structTagToDenom(data['l1_token']),
                l2StructTag: `0x1::native_${l2Denom}::Coin`,
                l2Denom
              };

              await transactionalEntityManager
                .getRepository(ChallengerCoinEntity)
                .save(coin);
              break;
            }
            case '0x1::op_bridge::TokenBridgeInitiatedEvent': {
              // handle token bridge initiated event
              const denom = `l2_${data['l2_token']}`;

              const entity: DepositTxEntity = {
                sequence: Number.parseInt(data['l1_sequence']),
                sender: data['from'],
                receiver: data['to'],
                amount: Number.parseInt(data['amount']),
                outputIndex: lastIndex + 1,
                finalizedOutputIndex: null,
                coinType: denom,
                isChecked: false
              };

              await transactionalEntityManager
                .getRepository(DepositTxEntity)
                .save(entity);
              break;
            }
          }
        }
      }
    );
  }
}
