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

    this.logger.info(`Registering ${l2Denom}...`);

    await this.helper.saveEntity(manager, ExecutorCoinEntity, coin);

    return new MsgCreateToken(
      wallet.key.accAddress,
      l2Denom,
      coinInfo.name,
      coinInfo.symbol,
      coinInfo.decimals
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

    const denom = `l2_${data['l2_token']}`;
    const entity: ExecutorDepositTxEntity = {
      sequence: Number.parseInt(data['l1_sequence']),
      sender: data['from'],
      receiver: data['to'],
      amount: Number.parseInt(data['amount']),
      outputIndex: lastIndex + 1,
      coinType: denom
    };

    this.logger.info(`Deposit tx in height ${this.syncedHeight}`);
    await manager.getRepository(ExecutorDepositTxEntity).save(entity);

    return new MsgDeposit(
      wallet.key.accAddress,
      AccAddress.fromHex(data['from']),
      AccAddress.fromHex(data['to']),
      new Coin(denom, data['amount']),
      Number.parseInt(data['l1_sequence']),
      this.syncedHeight + 1,
      Buffer.from(data['data'])
    );
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
              const msg: MsgCreateToken = await this.handleTokenRegisteredEvent(
                wallet,
                transactionalEntityManager,
                data
              );
              msgs.push(msg);
              break;
            }
            case '0x1::op_bridge::TokenBridgeInitiatedEvent': {
              // handle token bridge initiated event
              const msg: MsgDeposit =
                await this.handleTokenBridgeInitiatedEvent(
                  wallet,
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
          await wallet
            .transaction(msgs)
            .then((info) => {
              this.logger.info(
                `Succeed to submit tx in height: ${this.syncedHeight}\ntxhash: ${info?.txhash}\nmsgs: ${stringfyMsgs}`
              );
            })
            .catch(async (err) => {
              const errMsg = err.response?.data
                ? JSON.stringify(err.response?.data)
                : err;
              this.logger.error(
                `Failed to submit tx in height: ${this.syncedHeight}\n${stringfyMsgs}\n${errMsg}`
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
