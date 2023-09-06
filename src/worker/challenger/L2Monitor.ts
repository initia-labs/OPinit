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
import { sha3_256 } from 'lib/util';
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
          const lastOutput = await this.helper.getLastOutputFromDB(
            transactionalEntityManager,
            ChallengerOutputEntity
          );

          const lastCheckpointBlockHeight =
            lastOutput.length == 0 ? 0 : lastOutput[0].checkpointBlockHeight;

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
    coin: ChallengerCoinEntity,
    lastIndex: number
  ): ChallengerWithdrawalTxEntity {
    return {
      sequence: Number.parseInt(data['l2_sequence']),
      sender: data['from'],
      receiver: data['to'],
      amount: Number.parseInt(data['amount']),
      l2Id: data['l2_id'],
      coinType: coin.l1StructTag,
      outputIndex: lastIndex + 1,
      merkleRoot: '',
      merkleProof: [],
      isChecked: false
    };
  }

  public async handleEvents(): Promise<void> {
    await this.db.transaction(
      async (transactionalEntityManager: EntityManager) => {
        const lastIndex = await this.helper.getLastOutputIndex(
          transactionalEntityManager,
          ChallengerOutputEntity
        );

        const events = await this.helper.fetchEvents(
          config.l2lcd,
          this.syncedHeight
        );

        for (const evt of events) {
          const attrMap = this.helper.eventsToAttrMap(evt);

          switch (attrMap['type_tag']) {
            case '0x1::op_bridge::TokenBridgeInitiatedEvent': {
              const data: { [key: string]: string } =
                this.helper.parseData(attrMap);
              const l2Denom = data['l2_token'].replace('native_', '');
              const coin = await this.helper.getCoin(
                transactionalEntityManager,
                ChallengerCoinEntity,
                l2Denom
              );

              if (!coin) {
                this.logger.warn(`coin not found: ${l2Denom}`);
                continue;
              }

              const tx: ChallengerWithdrawalTxEntity = this.genTx(
                data,
                coin,
                lastIndex
              );

              this.logger.info(
                `withdraw tx found in output index : ${tx.outputIndex}`
              );
              await this.helper.saveEntity(
                transactionalEntityManager,
                ChallengerWithdrawalTxEntity,
                tx
              );
              break;
            }
            case '0x1::op_bridge::TokenBridgeFinalizedEvent': {
              const data: { [key: string]: string } =
                this.helper.parseData(attrMap);

              const l2Denom = data['l2_token'].replace('native_', '');

              const depositTx = await this.helper.getDepositTx(
                transactionalEntityManager,
                ChallengerDepositTxEntity,
                Number.parseInt(data['l1_sequence']),
                l2Denom
              );
              if (!depositTx) continue;

              const lastIndex = await this.helper.getLastOutputIndex(
                transactionalEntityManager,
                ChallengerOutputEntity
              );
              const isTxSame = (
                originTx: ChallengerDepositTxEntity
              ): boolean => {
                return (
                  originTx.sequence === Number.parseInt(data['l1_sequence']) &&
                  originTx.sender === data['from'] &&
                  originTx.receiver === data['to'] &&
                  originTx.amount === Number.parseInt(data['amount'])
                );
              };
              const finalizedIndex = isTxSame(depositTx) ? lastIndex + 1 : null;

              await transactionalEntityManager
                .getRepository(ChallengerDepositTxEntity)
                .update(
                  {
                    sequence: depositTx.sequence,
                    coinType: depositTx.coinType
                  },
                  { finalizedOutputIndex: finalizedIndex }
                );
              break;
            }
          }
        }
      }
    );
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
        const txEntities = await transactionalEntityManager
          .getRepository(ChallengerWithdrawalTxEntity)
          .find({
            where: { outputIndex: lastIndex + 1 }
          });

        const txs: WithdrawalTx[] = txEntities.map((entity) => ({
          sequence: entity.sequence,
          sender: entity.sender,
          receiver: entity.receiver,
          amount: entity.amount,
          l2_id: entity.l2Id,
          coin_type: entity.coinType
        }));
        const storage = new WithdrawalStorage(txs);
        const storageRoot = storage.getMerkleRoot();

        // save merkle root and proof for each tx
        for (const entity of txEntities) {
          const tx: WithdrawalTx = {
            sequence: entity.sequence,
            sender: entity.sender,
            receiver: entity.receiver,
            amount: entity.amount,
            l2_id: entity.l2Id,
            coin_type: entity.coinType
          };

          entity.merkleRoot = storageRoot;
          entity.merkleProof = storage.getMerkleProof(tx);
          await transactionalEntityManager
            .getRepository(ChallengerWithdrawalTxEntity)
            .save(entity);
        }

        // get output root and save to db
        const version = lastIndex + 1;
        const stateRoot = blockInfo.block.header.app_hash;
        const lastBlockHash = blockInfo.block_id.hash;

        const outputRoot = sha3_256(
          Buffer.concat([
            Buffer.from(version.toString()),
            Buffer.from(stateRoot, 'base64'),
            Buffer.from(storageRoot, 'hex'),
            Buffer.from(lastBlockHash, 'base64')
          ])
        ).toString('hex');

        const outputEntity: ChallengerOutputEntity = {
          outputIndex: lastIndex + 1,
          outputRoot,
          stateRoot,
          storageRoot,
          lastBlockHash,
          checkpointBlockHeight:
            this.nextCheckpointBlockHeight - this.submissionInterval // start block height of the epoch
        };

        await transactionalEntityManager
          .getRepository(ChallengerOutputEntity)
          .save(outputEntity);
        this.nextCheckpointBlockHeight += this.submissionInterval;
      }
    );
  }
}
