import { Monitor } from 'worker/bridgeExecutor/Monitor';
import { structTagToDenom } from 'lib/util';
import {
  ChallengerCoinEntity,
  ChallengerDepositTxEntity,
  ChallengerOutputEntity
} from 'orm';
import { getCoinInfo } from 'lib/lcd';
import { EntityManager } from 'typeorm';
import { RPCSocket } from 'lib/rpc';
import { getDB } from './db';
import winston from 'winston';
import { getConfig } from 'config';

const config = getConfig();

export class L1Monitor extends Monitor {
  constructor(public socket: RPCSocket, logger: winston.Logger) {
    super(socket, logger);
    [this.db] = getDB();
  }

  public name(): string {
    return 'challenger_l1_monitor';
  }

  public async handleTokenRegisteredEvent(
    manager: EntityManager,
    data: { [key: string]: string }
  ) {
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

    await this.helper.saveEntity(manager, ChallengerCoinEntity, coin);
  }

  public async handleTokenBridgeInitiatedEvent(
    manager: EntityManager,
    data: { [key: string]: string }
  ) {
    const lastIndex = await this.helper.getLastOutputIndex(
      manager,
      ChallengerOutputEntity
    );
     
    const denom = `l2_${data['l2_token']}`;
    const entity: ChallengerDepositTxEntity = {
      sequence: Number.parseInt(data['l1_sequence']),
      sender: data['from'],
      receiver: data['to'],
      amount: Number.parseInt(data['amount']),
      outputIndex: lastIndex + 1,
      height: this.syncedHeight,
      finalizedOutputIndex: null,
      coinType: denom,
      isChecked: false
    };

    await manager.getRepository(ChallengerDepositTxEntity).save(entity);
  }

  public async handleEvents(): Promise<void> {
    await this.db.transaction(
      async (transactionalEntityManager: EntityManager) => {
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
              await this.handleTokenRegisteredEvent(
                transactionalEntityManager,
                data
              );
              break;
            }
            case '0x1::op_bridge::TokenBridgeInitiatedEvent': {
              await this.handleTokenBridgeInitiatedEvent(
                transactionalEntityManager,
                data
              );
              break;
            }
          }
        }
      }
    );
  }
}
