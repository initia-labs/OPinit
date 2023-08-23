import config from 'config';
import {
  ChallengerCoinEntity,
  ChallengerOutputEntity,
  DepositTxEntity,
  WithdrawalTxEntity
} from 'orm';
import { Monitor } from './Monitor';
import { fetchBridgeConfig } from 'lib/lcd';
import { WithdrawalStorage } from 'lib/storage';
import { BridgeConfig, WithdrawalTx } from 'lib/types';
import { sha3_256 } from 'lib/util';
import { logger } from 'lib/logger';
import { EntityManager } from 'typeorm';

export class L2Monitor extends Monitor {
  submissionInterval: number;
  nextCheckpointBlockHeight: number;

  public name(): string {
    return 'challenger_l2_monitor';
  }

  public color(): string {
    return 'green';
  }

  public async run(): Promise<void> {
    try {
      await this.db.transaction(
        async (transactionalEntityManager: EntityManager) => {
          const lastOutput = await this.getLastOutputFromDB(
            transactionalEntityManager
          );

          const lastCheckpointBlockHeight =
            lastOutput.length == 0 ? 0 : lastOutput[0].checkpointBlockHeight;

          const cfg: BridgeConfig = await fetchBridgeConfig();
          this.submissionInterval = parseInt(cfg.submission_interval);

          const checkpointBlockHeight =
            lastCheckpointBlockHeight === 0
              ? parseInt(cfg.starting_block_number)
              : lastCheckpointBlockHeight + this.submissionInterval;

          this.nextCheckpointBlockHeight =
            checkpointBlockHeight + this.submissionInterval;
        }
      );
      await super.run();
    } catch (e) {
      logger.error('L2Monitor runs error:', e);
    }
  }

  public async handleEvents(): Promise<void> {
    await this.db.transaction(
      async (transactionalEntityManager: EntityManager) => {
        const lastIndex = await this.getLastOutputIndex(
          transactionalEntityManager
        );

        const searchRes = await config.l2lcd.tx.search({
          events: [
            { key: 'tx.height', value: (this.syncedHeight + 1).toString() }
          ]
        });

        const events = searchRes.txs
          .flatMap((tx) => tx.logs ?? [])
          .flatMap((log) => log.events);

        for (const evt of events) {
          if (evt.type !== 'move') continue;

          const attrMap: { [key: string]: string } = evt.attributes.reduce(
            (obj, attr) => {
              obj[attr.key] = attr.value;
              return obj;
            },
            {}
          );

          switch (attrMap['type_tag']) {
            case '0x1::op_bridge::TokenBridgeInitiatedEvent': {
              const data: { [key: string]: string } = JSON.parse(
                attrMap['data']
              );
              const l2Denom = data['l2_token'].replace('native_', '');
              const coin = await transactionalEntityManager
                .getRepository(ChallengerCoinEntity)
                .findOne({
                  where: { l2Denom }
                });

              if (!coin) {
                logger.warn(`coin not found: ${l2Denom}`);
                continue;
              }

              const tx: WithdrawalTxEntity = {
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

              logger.info(
                `withdraw tx found in output index : ${tx.outputIndex}`
              );

              await transactionalEntityManager
                .getRepository(WithdrawalTxEntity)
                .save(tx);
              break;
            }
            case '0x1::op_bridge::TokenBridgeFinalizedEvent': {
              const data: { [key: string]: string } = JSON.parse(
                attrMap['data']
              );
              const l2Denom = data['l2_token'].replace('native_', '');

              // get unchecked deposit tx
              const depositTx = await this.getDepositTx(
                transactionalEntityManager,
                Number.parseInt(data['l1_sequence']),
                l2Denom
              );

              if (!depositTx) continue;

              const lastIndex = await this.getLastOutputIndex(
                transactionalEntityManager
              );
              const isTxSame = (originTx: DepositTxEntity): boolean => {
                return (
                  originTx.sequence === Number.parseInt(data['l1_sequence']) &&
                  originTx.sender === data['from'] &&
                  originTx.receiver === data['to'] &&
                  originTx.amount === Number.parseInt(data['amount'])
                );
              };
              const finalizedIndex = isTxSame(depositTx) ? lastIndex + 1 : null;
              await transactionalEntityManager
                .getRepository(DepositTxEntity)
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

  async getDepositTx(
    transactionalEntityManager: EntityManager,
    sequence: number,
    coinType: string
  ): Promise<DepositTxEntity | null> {
    return await transactionalEntityManager
      .getRepository(DepositTxEntity)
      .findOne({
        where: { sequence, coinType }
      });
  }

  public async handleBlock(): Promise<void> {
    await this.db.transaction(
      async (transactionalEntityManager: EntityManager) => {
        if (this.syncedHeight < this.nextCheckpointBlockHeight - 1) {
          return;
        }

        const lastIndex = await this.getLastOutputIndex(
          transactionalEntityManager
        );
        const blockInfo = await config.l2lcd.tendermint.blockInfo(
          this.syncedHeight
        );

        // fetch txs and build merkle tree for withdrawal storage
        const txEntities = await transactionalEntityManager
          .getRepository(WithdrawalTxEntity)
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
            .getRepository(WithdrawalTxEntity)
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
