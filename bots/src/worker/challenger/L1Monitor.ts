import { Monitor } from 'worker/bridgeExecutor/Monitor';
import {
  ChallengerDepositTxEntity,
  ChallengerFinalizeWithdrawalTxEntity
} from 'orm';
import { EntityManager } from 'typeorm';
import { RPCClient, RPCSocket } from 'lib/rpc';
import { getDB } from './db';
import winston from 'winston';
import { getConfig } from 'config';

const config = getConfig();

export class L1Monitor extends Monitor {
  constructor(
    public socket: RPCSocket,
    public rpcClient: RPCClient,
    logger: winston.Logger
  ) {
    super(socket, rpcClient, logger);
    [this.db] = getDB();
  }

  public name(): string {
    return 'challenger_l1_monitor';
  }

  public async handleInitiateTokenDeposit(
    manager: EntityManager,
    data: { [key: string]: string }
  ): Promise<void> {
    const entity: ChallengerDepositTxEntity = {
      sequence: data['l1_sequence'],
      sender: data['from'],
      receiver: data['to'],
      l1Denom: data['l1_denom'],
      l2Denom: data['l2_denom'],
      amount: data['amount'],
      data: data['data']
    };
    await manager.getRepository(ChallengerDepositTxEntity).save(entity);
  }

  public async handleFinalizeTokenWithdrawalEvent(
    manager: EntityManager,
    data: { [key: string]: string }
  ): Promise<void> {
    const entity: ChallengerFinalizeWithdrawalTxEntity = {
      bridgeId: data['bridge_id'],
      outputIndex: parseInt(data['output_index']),
      sequence: data['l2_sequence'],
      sender: data['from'],
      receiver: data['to'],
      l1Denom: data['l1_denom'],
      l2Denom: data['l2_denom'],
      amount: data['amount']
    };

    await manager
      .getRepository(ChallengerFinalizeWithdrawalTxEntity)
      .save(entity);
  }

  public async handleEvents(manager: EntityManager): Promise<boolean> {
    const [isEmpty, depositEvents] = await this.helper.fetchEvents(
      config.l1lcd,
      this.currentHeight,
      'initiate_token_deposit'
    );

    if (isEmpty) return false;

    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const [_, withdrawalEvents] = await this.helper.fetchEvents(
      config.l1lcd,
      this.currentHeight,
      'finalize_token_withdrawal'
    );

    for (const evt of depositEvents) {
      const attrMap = this.helper.eventsToAttrMap(evt);
      if (attrMap['bridge_id'] !== this.bridgeId.toString()) continue;
      await this.handleInitiateTokenDeposit(manager, attrMap);
    }

    for (const evt of withdrawalEvents) {
      const attrMap = this.helper.eventsToAttrMap(evt);
      if (attrMap['bridge_id'] !== this.bridgeId.toString()) continue;
      await this.handleFinalizeTokenWithdrawalEvent(manager, attrMap);
    }
    return true;
  }
}
