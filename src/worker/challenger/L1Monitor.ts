import { Monitor } from 'worker/bridgeExecutor/Monitor';
import {
  ChallengerCoinEntity,
  ChallengerDepositTxEntity,
  ChallengerOutputEntity
} from 'orm';
import {
  CoinInfo,
  computeCoinMetadata,
  normalizeMetadata,
  resolveFAMetadata
} from 'lib/lcd';
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
    const l1Metadata = data['l1_token'];
    const l2Metadata = normalizeMetadata(
      computeCoinMetadata('0x1', 'l2/' + data['l2_token'])
    );

    const l1CoinInfo: CoinInfo = await resolveFAMetadata(
      config.l1lcd,
      l1Metadata
    );

    const l1Denom = l1CoinInfo.denom;
    const l2Denom = 'l2/' + data['l2_token'];

    const coin: ChallengerCoinEntity = {
      l1Metadata: l1Metadata,
      l1Denom: l1Denom,
      l2Metadata: l2Metadata,
      l2Denom: l2Denom
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

    const l2Metadata = normalizeMetadata(
      computeCoinMetadata('0x1', 'l2/' + data['l2_token'])
    );

    const entity: ChallengerDepositTxEntity = {
      sequence: Number.parseInt(data['l1_sequence']),
      sender: data['from'],
      receiver: data['to'],
      amount: Number.parseInt(data['amount']),
      outputIndex: lastIndex + 1,
      metadata: l2Metadata,
      height: this.syncedHeight,
      finalizedOutputIndex: null,
      isChecked: false
    };

    await manager.getRepository(ChallengerDepositTxEntity).save(entity);
  }

  public async handleEvents(): Promise<void> {
    await this.db.transaction(
      async (transactionalEntityManager: EntityManager) => {
        const events = await this.helper.fetchEvents(
          config.l1lcd,
          this.syncedHeight,
          'move'
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
