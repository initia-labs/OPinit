import {
  ChallengerCoinEntity,
  ChallengerOutputEntity,
  ChallengerDepositTxEntity,
  ChallengerWithdrawalTxEntity
} from 'orm';
import { Monitor } from 'worker/bridgeExecutor/Monitor';
import { fetchBridgeConfig } from 'lib/lcd';
import { WithdrawalStorage } from 'lib/storage';
import { BridgeConfig, WithdrawalTx } from 'lib/types';
import { EntityManager } from 'typeorm';
import { RPCSocket } from 'lib/rpc';
import winston from 'winston';
import { getDB } from './db';
import { getConfig } from 'config';

const config = getConfig();

export class L2Monitor extends Monitor {
  submissionInterval: number;
  nextCheckpointBlockHeight: number;

  constructor(public socket: RPCSocket, logger: winston.Logger) {
    super(socket, logger);
    [this.db] = getDB();
  }

  public name(): string {
    return 'challenger_l2_monitor';
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
              ChallengerOutputEntity
            );
          await this.configureBridge(lastCheckpointBlockHeight);
          await super.run();
        }
      );
    } catch (err) {
      throw new Error(err);
    }
  }

  private genTx(
    data: { [key: string]: string },
    coin: ChallengerCoinEntity,
    lastIndex: number
  ): ChallengerWithdrawalTxEntity {
    return {
      sequence: Number.parseInt(data['l2_sequence']),
      sender: data['from'],
      receiver: data['to'],
      amount: Number.parseInt(data['amount']),
      l2Id: config.L2ID,
      coinType: coin.l1StructTag,
      outputIndex: lastIndex + 1,
      merkleRoot: '',
      merkleProof: [],
      isChecked: false
    };
  }

  private async handleTokenBridgeInitiatedEvent(
    manager: EntityManager,
    data: { [key: string]: string }
  ) {
    const lastIndex = await this.helper.getLastOutputIndex(
      manager,
      ChallengerOutputEntity
    );

    const l2Denom = data['l2_token'].replace('native_', '');
    const coin = await this.helper.getCoin(
      manager,
      ChallengerCoinEntity,
      l2Denom
    );

    if (!coin) {
      this.logger.warn(`coin not found for ${l2Denom}`);
      return;
    }

    const tx: ChallengerWithdrawalTxEntity = this.genTx(data, coin, lastIndex);
    this.logger.info(`withdraw tx in height ${this.syncedHeight}`);
    await this.helper.saveEntity(manager, ChallengerWithdrawalTxEntity, tx);
  }

  private async handleTokenBridgeFinalizedEvent(
    manager: EntityManager,
    data: { [key: string]: string }
  ) {
    const l2Denom = data['l2_token'].replace('native_', '');
    const depositTx = await this.helper.getDepositTx(
      manager,
      ChallengerDepositTxEntity,
      Number.parseInt(data['l1_sequence']),
      l2Denom
    );
    if (!depositTx) return;

    const lastIndex = await this.helper.getLastOutputIndex(
      manager,
      ChallengerOutputEntity
    );

    const isTxSame = (originTx: ChallengerDepositTxEntity): boolean => {
      return (
        originTx.sequence === Number.parseInt(data['l1_sequence']) &&
        originTx.sender === data['from'] &&
        originTx.receiver === data['to'] &&
        originTx.amount === Number.parseInt(data['amount'])
      );
    };
    const finalizedIndex = isTxSame(depositTx) ? lastIndex + 1 : null;

    await manager.getRepository(ChallengerDepositTxEntity).update(
      {
        sequence: depositTx.sequence,
        coinType: depositTx.coinType
      },
      { finalizedOutputIndex: finalizedIndex }
    );
  }

  public async handleEvents(): Promise<void> {
    await this.db.transaction(
      async (transactionalEntityManager: EntityManager) => {
        const events = await this.helper.fetchEvents(
          config.l2lcd,
          this.syncedHeight
        );

        for (const evt of events) {
          const attrMap = this.helper.eventsToAttrMap(evt);
          const data: { [key: string]: string } =
            this.helper.parseData(attrMap);

          switch (attrMap['type_tag']) {
            case '0x1::op_bridge::TokenBridgeInitiatedEvent': {
              await this.handleTokenBridgeInitiatedEvent(
                transactionalEntityManager,
                data
              );
              break;
            }
            case '0x1::op_bridge::TokenBridgeFinalizedEvent': {
              await this.handleTokenBridgeFinalizedEvent(
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

  private async saveMerkleRootAndProof(
    manager: EntityManager,
    entities: ChallengerWithdrawalTxEntity[]
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
        ChallengerWithdrawalTxEntity,
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
          ChallengerOutputEntity
        );
        const blockInfo = await config.l2lcd.tendermint.blockInfo(
          this.syncedHeight
        );

        // fetch txs and build merkle tree for withdrawal storage
        const txEntities = await this.helper.getWithdrawalTxs(
          transactionalEntityManager,
          ChallengerWithdrawalTxEntity,
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
          ChallengerOutputEntity,
          outputEntity
        );
        this.nextCheckpointBlockHeight += this.submissionInterval;
      }
    );
  }
}
