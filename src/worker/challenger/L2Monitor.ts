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
import { getL2Denom, sha3_256 } from 'lib/util';
import { logger } from 'lib/logger';

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
      const lastOutput = await this.getLastOutputFromDB();
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

      await super.run();
    } catch (e) {
      logger.error('L2Monitor runs error:', e);
    }
  }

  public async handleEvents(): Promise<void> {
    const lastOutput = await this.getLastOutputFromDB();
    const lastIndex = lastOutput.length == 0 ? -1 : lastOutput[0].outputIndex;

    const searchRes = await config.l2lcd.tx.search({
      events: [{ key: 'tx.height', value: (this.syncedHeight + 1).toString() }]
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
          const data: { [key: string]: string } = JSON.parse(attrMap['data']);
          const l2Denom = Buffer.from(data['l2_token'])
            .toString()
            .replace('native_', '');
          const coin = await this.db
            .getRepository(ChallengerCoinEntity)
            .findOne({
              where: { l2Denom }
            });

          if (!coin) continue;

          const tx: WithdrawalTxEntity = {
            sequence: Number.parseInt(data['l2_sequence']),
            sender: data['from'],
            receiver: data['to'],
            amount: Number.parseInt(data['amount']),
            coinType: coin.l1StructTag,
            outputIndex: lastIndex + 1,
            merkleRoot: '',
            merkleProof: [],
            isChecked: false
          };

          logger.info(`withdraw tx found in output index : ${tx.outputIndex}`);

          await this.db.getRepository(WithdrawalTxEntity).save(tx);
          break;
        }
        case '0x1::op_bridge::TokenBridgeFinalizedEvent': {
          const data: { [key: string]: string } = JSON.parse(attrMap['data']);
          const l2Denom = Buffer.from(data['l2_token'])
            .toString()
            .replace('native_', '');

          // get unchecked deposit tx
          const depositTx = await this.getDepositTx(
            Number.parseInt(data['l1_sequence']),
            l2Denom
          );

          // challenger could miss to store deposit tx due to `fetchBridgeState()`
          if (depositTx === null) return;

          const lastIndex = await this.getLastOutputIndex();
          // compare initiate and finalize tx
          const originTx = depositTx;
          const finalTx = {
            sequence: Number.parseInt(data['l1_sequence']),
            sender: data['from'],
            receiver: data['to'],
            amount: Number.parseInt(data['amount'])
          };
          const isTxSame = (originTx, finalTx): boolean => {
            return (
              originTx.sequence === finalTx.sequence &&
              originTx.sender === finalTx.sender &&
              originTx.receiver === finalTx.receiver &&
              originTx.amount === finalTx.amount
            );
          };
          const index = isTxSame(originTx, finalTx) ? lastIndex + 1 : -1;
          await this.db.getRepository(DepositTxEntity).save({
            ...depositTx,
            finalizedOutputIndex: index
          });
          break;
        }
      }
    }
  }

  async getDepositTx(
    sequence: number,
    coinType: string
  ): Promise<DepositTxEntity | null> {
    return await this.db.getRepository(DepositTxEntity).findOne({
      where: { sequence, coinType }
    });
  }

  public async handleBlock(): Promise<void> {
    if (this.syncedHeight < this.nextCheckpointBlockHeight - 1) {
      return;
    }

    const lastOutput = await this.db
      .getRepository(ChallengerOutputEntity)
      .find({
        order: { outputIndex: 'DESC' },
        take: 1
      });
    const lastIndex = lastOutput.length == 0 ? -1 : lastOutput[0].outputIndex;
    const blockInfo = await config.l2lcd.tendermint.blockInfo(
      this.syncedHeight
    );

    // fetch txs and build merkle tree for withdrawal storage
    const txEntities = await this.db.getRepository(WithdrawalTxEntity).find({
      where: { outputIndex: lastIndex + 1 }
    });

    const txs: WithdrawalTx[] = txEntities.map((entity) => ({
      sequence: entity.sequence,
      sender: entity.sender,
      receiver: entity.receiver,
      amount: entity.amount,
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
        coin_type: entity.coinType
      };

      entity.merkleRoot = storageRoot;
      entity.merkleProof = storage.getMerkleProof(tx);
      await this.db.getRepository(WithdrawalTxEntity).save(entity);
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

    await this.db.getRepository(ChallengerOutputEntity).save(outputEntity);
    this.nextCheckpointBlockHeight += this.submissionInterval;
  }
}
