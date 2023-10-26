import { Monitor } from './Monitor';
import {
  CoinInfo,
  computeCoinMetadata,
  normalizeMetadata,
  resolveFAMetadata
} from 'lib/lcd';
import {
  AccAddress,
  Coin,
  Msg,
  MsgCreateToken,
  MsgDeposit
} from '@initia/minitia.js';

import {
  ExecutorCoinEntity,
  ExecutorDepositTxEntity,
  ExecutorFailedTxEntity,
  ExecutorOutputEntity
} from 'orm';
import { WalletType, getWallet, TxWallet } from 'lib/wallet';
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
    return 'executor_l1_monitor';
  }

  public async handleTokenRegisteredEvent(
    wallet: TxWallet,
    manager: EntityManager,
    data: { [key: string]: string }
  ): Promise<MsgCreateToken> {
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

    const coin: ExecutorCoinEntity = {
      l1Metadata: l1Metadata,
      l1Denom: l1Denom,
      l2Metadata: l2Metadata,
      l2Denom: l2Denom,
      isChecked: false
    };

    await this.helper.saveEntity(manager, ExecutorCoinEntity, coin);

    return new MsgCreateToken(
      wallet.key.accAddress,
      l1CoinInfo.name,
      l2Denom,
      l1CoinInfo.decimals
    );
  }

  public async handleTokenBridgeInitiatedEvent(
    wallet: TxWallet,
    manager: EntityManager,
    data: { [key: string]: string }
  ): Promise<MsgDeposit> {
    const lastIndex = await this.helper.getLastOutputIndex(
      manager,
      ExecutorOutputEntity
    );

    const l2Metadata = normalizeMetadata(
      computeCoinMetadata('0x1', 'l2/' + data['l2_token'])
    );
    const l2Denom = 'l2/' + data['l2_token'];

    const entity: ExecutorDepositTxEntity = {
      sequence: Number.parseInt(data['l1_sequence']),
      sender: data['from'],
      receiver: data['to'],
      amount: Number.parseInt(data['amount']),
      outputIndex: lastIndex + 1,
      metadata: l2Metadata,
      height: this.syncedHeight
    };

    this.logger.info(`Deposit tx in height ${this.syncedHeight}`);
    await manager.getRepository(ExecutorDepositTxEntity).save(entity);

    return new MsgDeposit(
      wallet.key.accAddress,
      AccAddress.fromHex(data['from']),
      AccAddress.fromHex(data['to']),
      new Coin(l2Denom, data['amount']),
      Number.parseInt(data['l1_sequence']),
      this.syncedHeight,
      Buffer.from(data['data'])
    );
  }

  public async handleEvents(): Promise<void> {
    await this.db.transaction(
      async (transactionalEntityManager: EntityManager) => {
        const msgs: Msg[] = [];
        const executor: TxWallet = getWallet(WalletType.Executor);

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
              const msg: MsgCreateToken = await this.handleTokenRegisteredEvent(
                executor,
                transactionalEntityManager,
                data
              );
              msgs.push(msg);
              break;
            }
            case '0x1::op_bridge::TokenBridgeInitiatedEvent': {
              const msg: MsgDeposit =
                await this.handleTokenBridgeInitiatedEvent(
                  executor,
                  transactionalEntityManager,
                  data
                );
              msgs.push(msg);
              break;
            }
          }
        }

        if (msgs.length > 0) {
          const stringfyMsgs = msgs.map((msg) => msg.toJSON().toString());
          await executor
            .transaction(msgs)
            .then((info) => {
              this.logger.info(
                `succeed to submit tx in height: ${this.syncedHeight}\ntxhash: ${info?.txhash}\nmsgs: ${stringfyMsgs}`
              );
            })
            .catch(async (err) => {
              const errMsg = err.response?.data
                ? JSON.stringify(err.response?.data)
                : err;
              this.logger.error(
                `Failed to submit tx in height: ${this.syncedHeight}\nMsg: ${stringfyMsgs}\n${errMsg}`
              );
              const failedTx: ExecutorFailedTxEntity = {
                height: this.syncedHeight,
                monitor: this.name(),
                messages: stringfyMsgs,
                error: errMsg
              };
              await this.helper.saveEntity(
                transactionalEntityManager,
                ExecutorFailedTxEntity,
                failedTx
              );
            });
        }
      }
    );
  }
}
