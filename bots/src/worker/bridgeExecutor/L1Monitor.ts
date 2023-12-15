import { Monitor } from './Monitor';
import { Coin, Msg, MsgFinalizeTokenDeposit, Wallet } from '@initia/initia.js';
import {
  ExecutorDepositTxEntity,
  ExecutorFailedTxEntity,
  ExecutorOutputEntity
} from 'orm';
import { EntityManager } from 'typeorm';
import { RPCClient, RPCSocket } from 'lib/rpc';
import { getDB } from './db';
import winston from 'winston';
import { config } from 'config';
import { TxWallet, WalletType, getWallet, initWallet } from 'lib/wallet';

export class L1Monitor extends Monitor {
  executor: TxWallet;

  constructor(
    public socket: RPCSocket,
    public rpcClient: RPCClient,
    logger: winston.Logger
  ) {
    super(socket, rpcClient, logger);
    [this.db] = getDB();
    initWallet(WalletType.Executor, config.l2lcd);
    this.executor = getWallet(WalletType.Executor);
  }

  public name(): string {
    return 'executor_l1_monitor';
  }

  public async handleInitiateTokenDeposit(
    wallet: Wallet,
    manager: EntityManager,
    data: { [key: string]: string }
  ): Promise<[ExecutorDepositTxEntity, MsgFinalizeTokenDeposit]> {
    const lastIndex = await this.helper.getLastOutputIndex(
      manager,
      ExecutorOutputEntity
    );

    const entity: ExecutorDepositTxEntity = {
      sequence: data['l1_sequence'],
      sender: data['from'],
      receiver: data['to'],
      l1Denom: data['l1_denom'],
      l2Denom: data['l2_denom'],
      amount: data['amount'],
      data: data['data'],
      outputIndex: lastIndex + 1,
      bridgeId: this.bridgeId.toString(),
      l1Height: this.currentHeight
    };

    return [
      entity,
      new MsgFinalizeTokenDeposit(
        wallet.key.accAddress,
        data['from'],
        data['to'],
        new Coin(data['l2_denom'], data['amount']),
        parseInt(data['l1_sequence']),
        this.currentHeight,
        Buffer.from(data['data'], 'hex').toString('base64')
      )
    ];
  }

  public async handleEvents(manager: EntityManager): Promise<any> {
    const [isEmpty, depositEvents] = await this.helper.fetchEvents(
      config.l1lcd,
      this.currentHeight,
      'initiate_token_deposit'
    );

    if (isEmpty) return false;

    const msgs: Msg[] = [];
    const entities: ExecutorDepositTxEntity[] = [];

    for (const evt of depositEvents) {
      const attrMap = this.helper.eventsToAttrMap(evt);
      if (attrMap['bridge_id'] !== this.bridgeId.toString()) continue;
      const [entity, msg] = await this.handleInitiateTokenDeposit(
        this.executor,
        manager,
        attrMap
      );

      entities.push(entity);
      if (msg) msgs.push(msg);
    }

    await this.processMsgs(manager, msgs, entities);
    return true;
  }

  async processMsgs(
    manager: EntityManager,
    msgs: Msg[],
    entities: ExecutorDepositTxEntity[]
  ): Promise<void> {
    if (msgs.length == 0) return;
    const stringfyMsgs = msgs.map((msg) => msg.toJSON().toString());
    try {
      for (const entity of entities) {
        await this.helper.saveEntity(manager, ExecutorDepositTxEntity, entity);
      }
      await this.executor.transaction(msgs);
      this.logger.info(
        `Succeeded to submit tx in height: ${this.currentHeight} ${stringfyMsgs}`
      );
    } catch (err) {
      const errMsg = err.response?.data
        ? JSON.stringify(err.response?.data)
        : err.toString();
      this.logger.error(
        `Failed to submit tx in height: ${this.currentHeight}\nMsg: ${stringfyMsgs}\nError: ${errMsg}`
      );

      // Save all entities in a single batch operation, if possible
      for (const entity of entities) {
        await this.helper.saveEntity(manager, ExecutorFailedTxEntity, {
          ...entity,
          error: errMsg,
          processed: false
        });
      }
    }
  }
}
