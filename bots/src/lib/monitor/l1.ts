import { Monitor } from './monitor'
import { Coin, Msg, MsgFinalizeTokenDeposit } from '@initia/initia.js'
import {
  ExecutorDepositTxEntity,
  ExecutorUnconfirmedTxEntity,
  ExecutorOutputEntity
} from '../../orm'
import { EntityManager } from 'typeorm'
import { RPCClient, RPCSocket } from '../rpc'
import { getDB } from '../../worker/bridgeExecutor/db'
import winston from 'winston'
import { config } from '../../config'
import { TxWallet, WalletType, getWallet, initWallet } from '../wallet'

export class L1Monitor extends Monitor {
  executor: TxWallet

  constructor(
    public socket: RPCSocket,
    public rpcClient: RPCClient,
    logger: winston.Logger
  ) {
    super(socket, rpcClient, logger);
    [this.db] = getDB()
    initWallet(WalletType.Executor, config.l2lcd)
    this.executor = getWallet(WalletType.Executor)
  }

  public name(): string {
    return 'executor_l1_monitor'
  }

  public async handleInitiateTokenDeposit(
    manager: EntityManager,
    data: { [key: string]: string }
  ): Promise<[ExecutorDepositTxEntity, MsgFinalizeTokenDeposit]> {
    const lastIndex = await this.helper.getLastOutputIndex(
      manager,
      ExecutorOutputEntity
    )

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
    }

    return [
      entity,
      new MsgFinalizeTokenDeposit(
        this.executor.key.accAddress,
        data['from'],
        data['to'],
        new Coin(data['l2_denom'], data['amount']),
        parseInt(data['l1_sequence']),
        this.currentHeight,
        data['l1_denom'],
        Buffer.from(data['data'], 'hex').toString('base64')
      )
    ]
  }

  public async handleEvents(manager: EntityManager): Promise<any> {
    const [isEmpty, events] = await this.helper.fetchAllEvents(
      config.l1lcd,
      this.currentHeight
    )

    if (isEmpty) return false

    const msgs: Msg[] = []
    const depositEntities: ExecutorDepositTxEntity[] = []

    const depositEvents = events.filter(
      (evt) => evt.type === 'initiate_token_deposit'
    )
    for (const evt of depositEvents) {
      const attrMap = this.helper.eventsToAttrMap(evt)
      if (attrMap['bridge_id'] !== this.bridgeId.toString()) continue
      const [entity, msg] = await this.handleInitiateTokenDeposit(
        manager,
        attrMap
      )

      depositEntities.push(entity)
      if (msg) msgs.push(msg)
    }

    await this.processMsgs(manager, msgs, depositEntities)
    return true
  }

  async processMsgs(
    manager: EntityManager,
    msgs: Msg[],
    depositEntities: ExecutorDepositTxEntity[]
  ): Promise<void> {
    if (msgs.length == 0) return
    const stringfyMsgs = msgs.map((msg) => msg.toJSON().toString())
    try {
      for (const entity of depositEntities) {
        await this.helper.saveEntity(manager, ExecutorDepositTxEntity, entity)
      }

      await this.executor.transaction(msgs)
      this.logger.info(
        `Succeeded to submit tx in height: ${this.currentHeight} ${stringfyMsgs}`
      )
    } catch (err) {
      const errMsg = err.response?.data
        ? JSON.stringify(err.response?.data)
        : err.toString()
      this.logger.info(
        `Failed to submit tx in height: ${this.currentHeight}\nMsg: ${stringfyMsgs}\nError: ${errMsg}`
      )

      for (const entity of depositEntities) {
        await this.helper.saveEntity(manager, ExecutorUnconfirmedTxEntity, {
          ...entity,
          error: errMsg,
          processed: false
        })
      }
    }
  }
}
