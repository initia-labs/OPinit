import config from 'config'
import { CoinEntity, OutputEntity, TxEntity } from 'orm'
import { Monitor } from './Monitor'
import { fetchBridgeConfig } from 'lib/lcd'
import { WithdrawalStorage } from 'lib/storage'
import { BridgeConfig, WithdrawalTx } from 'lib/types'
import { sha3_256 } from 'lib/util'
import { logger } from 'lib/logger'

export class L2Monitor extends Monitor {
  submissionInterval: number
  nextBlockHeight: number
  startBlockHeight: number

  public name(): string {
    return 'l2_monitor'
  }

  public async run(): Promise<void> {
    try {
      const lastOutput = await this.getLastOutputFromDB();
      const lastStartBlockHeight = lastOutput[1] == 0 ? 0 : lastOutput[0][0].startBlockHeight;
      const cfg: BridgeConfig = await fetchBridgeConfig();
  
      this.startBlockHeight = parseInt(cfg.starting_block_number);
      this.submissionInterval = parseInt(cfg.submission_interval);
  
      this.nextBlockHeight = lastStartBlockHeight === 0
        ? this.startBlockHeight + this.submissionInterval
        : lastStartBlockHeight + this.submissionInterval;
  
      await super.run();
    } catch (e) {
      logger.error('L2Monitor runs error:', e);
    }
  }

  public async getLastOutputFromDB():Promise<[OutputEntity[], number]>{
    return await this.db.getRepository(OutputEntity).findAndCount({
      order: { outputIndex: 'DESC' },
      take: 1,
    })
  }
  public async handleEvents(): Promise<void> {
    const lastOutput = await this.getLastOutputFromDB()
    const lastIndex = lastOutput[1] == 0 ? -1 : lastOutput[0][0].outputIndex

    const searchRes = await config.l2lcd.tx.search({
      events: [
        { key: 'tx.height', value: (this.syncedHeight + 1).toString() },
      ],
    })
    const events = searchRes.txs.flatMap((tx) => tx.logs ?? []).flatMap((log) => log.events)

    for (const evt of events) {
      if (evt.type !== 'move') continue

      const attrMap: { [key: string]: string } = evt.attributes.reduce(
        (obj, attr) => {
          obj[attr.key] = attr.value
          return obj
        },
        {}
      )

      if (attrMap['type_tag'] !== '0x1::op_bridge::TokenBridgeInitiatedEvent') {
        continue
      }

      const data: { [key: string]: string } = JSON.parse(attrMap['data'])
      const l2Denom = Buffer.from(data['l2_token'])
        .toString()
        .replace('native_', '')
      const coin = await this.db.getRepository(CoinEntity).findOne({
        where: { l2Denom },
      })

      const tx: TxEntity = {
        sequence: Number.parseInt(data['l2_sequence']),
        sender: data['from'],
        receiver: data['to'],
        amount: Number.parseInt(data['amount']),
        coin_type: coin?.l1StructTag ?? '',
        outputIndex: lastIndex + 1,
        merkleRoot: '',
        merkleProof: [],
      }

      logger.info(`withdraw tx found ${tx}`)

      await this.db.getRepository(TxEntity).save(tx)
    }
  }

  public async handleBlock(): Promise<void> {
    if (this.syncedHeight < this.nextBlockHeight - 1) {
      return
    }

    const lastOutput = await this.db.getRepository(OutputEntity).findAndCount({
      order: { outputIndex: 'DESC' },
      take: 1,
    })
    const lastIndex = lastOutput[1] == 0 ? -1 : lastOutput[0][0].outputIndex
    const blockInfo = await config.l2lcd.tendermint.blockInfo(this.syncedHeight)

    // fetch txs and build merkle tree for withdrawal storage
    const txEntities = await this.db.getRepository(TxEntity).find({
      where: { outputIndex: lastIndex + 1 },
    })
    const txs: WithdrawalTx[] = txEntities.map((entity) => ({
      sequence: entity.sequence,
      sender: entity.sender,
      receiver: entity.receiver,
      amount: entity.amount,
      coin_type: entity.coin_type,
    }))
    const storage = new WithdrawalStorage(txs)
    const storageRoot = storage.getMerkleRoot()

    // save merkle root and proof for each tx
    for (const entity of txEntities) {
      const tx: WithdrawalTx = {
        sequence: entity.sequence,
        sender: entity.sender,
        receiver: entity.receiver,
        amount: entity.amount,
        coin_type: entity.coin_type,
      }

      entity.merkleRoot = storageRoot
      entity.merkleProof = storage.getMerkleProof(tx)
      await this.db.getRepository(TxEntity).save(entity)
    }

    // get output root and save to db
    const version = lastIndex + 1
    const stateRoot = blockInfo.block.header.app_hash
    const lastBlockHash = blockInfo.block_id.hash

    const outputRoot = sha3_256(
      Buffer.concat([
        Buffer.from(version.toString()),
        Buffer.from(stateRoot, 'base64'),
        Buffer.from(storageRoot, 'hex'),
        Buffer.from(lastBlockHash, 'base64'),
      ])
    ).toString('hex')
    const outputEntity: OutputEntity = {
      outputIndex: lastIndex + 1,
      outputRoot,
      stateRoot,
      storageRoot,
      lastBlockHash,
      startBlockHeight: this.nextBlockHeight - this.submissionInterval, // start block height of the epoch
    }
    
    await this.db.getRepository(OutputEntity).save(outputEntity)
    this.nextBlockHeight += this.submissionInterval
  }
}
