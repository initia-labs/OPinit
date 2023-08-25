import config from 'config';
import {
  ExecutorCoinEntity,
  ExecutorOutputEntity,
  ExecutorWithdrawalTxEntity
} from 'orm';
import { Monitor } from './Monitor';
import { fetchBridgeConfig } from 'lib/lcd';
import { WithdrawalStorage } from 'lib/storage';
import { BridgeConfig, WithdrawalTx } from 'lib/types';
import { EntityManager } from 'typeorm';
import { BlockInfo } from '@initia/minitia.js';
import { getDB } from './db';
import { RPCSocket } from 'lib/rpc';
import winston from 'winston';

export class L2Monitor extends Monitor {
  submissionInterval: number;
  nextCheckpointBlockHeight: number;

  constructor(public socket: RPCSocket, logger: winston.Logger) {
    super(socket, logger);
    [this.db] = getDB();
  }

  public name(): string {
    return 'executor_l2_monitor';
  }

  private async configureBridge(
    lastCheckpointBlockHeight: number
  ): Promise<void> {
    const cfg: BridgeConfig = await fetchBridgeConfig();
    this.submissionInterval = parseInt(cfg.submission_interval);

    const checkpointBlockHeight =
      lastCheckpointBlockHeight === 0
        ? parseInt(cfg.starting_block_number)
        : lastCheckpointBlockHeight + this.submissionInterval;

    this.nextCheckpointBlockHeight =
      checkpointBlockHeight + this.submissionInterval;
  }

  public async run(): Promise<void> {
    try {
      await this.db.transaction(
        async (transactionalEntityManager: EntityManager) => {
          const lastCheckpointBlockHeight =
            await this.helper.getCheckpointBlockHeight(
              transactionalEntityManager,
              ExecutorOutputEntity
            );
          await this.configureBridge(lastCheckpointBlockHeight);
          await super.run();
        }
      );
    } catch (err) {
      throw new Error(`Error in L2 Monitor ${err}`);
    }
  }

  private genTx(
    data: { [key: string]: string },
    coin: ExecutorCoinEntity,
    lastIndex: number
  ): ExecutorWithdrawalTxEntity {
    return {
      sequence: Number.parseInt(data['l2_sequence']),
      sender: data['from'],
      receiver: data['to'],
      amount: Number.parseInt(data['amount']),
      l2Id: config.L2ID,
      coinType: coin.l1StructTag,
      outputIndex: lastIndex + 1,
      merkleRoot: '',
      merkleProof: []
    };
  }

  public async handleEvents(): Promise<void> {
    await this.db.transaction(
      async (transactionalEntityManager: EntityManager) => {
        const lastIndex = await this.helper.getLastOutputIndex(
          transactionalEntityManager,
          ExecutorOutputEntity
        );
        const events = await this.helper.fetchEvents(
          config.l2lcd,
          this.syncedHeight
        );

        for (const evt of events) {
          const attrMap = this.helper.eventsToAttrMap(evt);
          if (
            attrMap['type_tag'] !== '0x1::op_bridge::TokenBridgeInitiatedEvent'
          )
            continue;
          const data: { [key: string]: string } =
            this.helper.parseData(attrMap);

          const l2Denom = data['l2_token'].replace('native_', '');
          const coin = await this.helper.getCoin(
            transactionalEntityManager,
            ExecutorCoinEntity,
            l2Denom
          );

          if (!coin) {
            this.logger.warn(`coin not found for ${l2Denom}`);
            continue;
          }

          const tx: ExecutorWithdrawalTxEntity = this.genTx(
            data,
            coin,
            lastIndex
          );

          this.logger.info(
            `withdraw tx found in output index : ${tx.outputIndex}`
          );
          await this.helper.saveEntity(
            transactionalEntityManager,
            ExecutorWithdrawalTxEntity,
            tx
          );
        }
      }
    );
  }

  private async saveMerkleRootAndProof(
    manager: EntityManager,
    entities: ExecutorWithdrawalTxEntity[]
  ): Promise<string> {
    const txs: WithdrawalTx[] = entities.map((entity) => ({
      sequence: entity.sequence,
      sender: entity.sender,
      receiver: entity.receiver,
      amount: entity.amount,
      l2_id: entity.l2Id,
      coin_type: entity.coinType
    }));

    const storage = new WithdrawalStorage(txs);
    const storageRoot = storage.getMerkleRoot();
    for (let i = 0; i < entities.length; i++) {
      entities[i].merkleRoot = storageRoot;
      entities[i].merkleProof = storage.getMerkleProof(txs[i]);
      await this.helper.saveEntity(
        manager,
        ExecutorWithdrawalTxEntity,
        entities[i]
      );
    }
    return storageRoot;
  }

  public async handleBlock(): Promise<void> {
    if (this.syncedHeight < this.nextCheckpointBlockHeight - 1) return;

    await this.db.transaction(
      async (transactionalEntityManager: EntityManager) => {
        const lastIndex = await this.helper.getLastOutputIndex(
          transactionalEntityManager,
          ExecutorOutputEntity
        );
        const blockInfo: BlockInfo = await config.l2lcd.tendermint.blockInfo(
          this.syncedHeight
        );

        // fetch txs and build merkle tree for withdrawal storage
        const txEntities = await this.helper.getWithdrawalTxs(
          transactionalEntityManager,
          ExecutorWithdrawalTxEntity,
          lastIndex
        );
        const storageRoot = await this.saveMerkleRootAndProof(
          transactionalEntityManager,
          txEntities
        );
        const outputEntity = this.helper.calculateOutputEntity(
          lastIndex,
          blockInfo,
          storageRoot,
          this.nextCheckpointBlockHeight - this.submissionInterval
        );

        await this.helper.saveEntity(
          transactionalEntityManager,
          ExecutorOutputEntity,
          outputEntity
        );
        this.nextCheckpointBlockHeight += this.submissionInterval;
      }
    );
  }
}
