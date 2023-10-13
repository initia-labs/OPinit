import { BlockInfo } from '@initia/minitia.js';
import { sha3_256 } from 'lib/util';
import { EntityManager, EntityTarget, ObjectLiteral } from 'typeorm';

class MonitorHelper {
  ///
  /// GET
  ///
  public async getSyncedState<T extends ObjectLiteral>(
    manager: EntityManager,
    entityClass: EntityTarget<T>,
    name: string
  ): Promise<T | null> {
    return await manager.getRepository(entityClass).findOne({
      where: { name: name } as any
    });
  }

  public async getWithdrawalTxs<T extends ObjectLiteral>(
    manager: EntityManager,
    entityClass: EntityTarget<T>,
    lastIndex: number
  ): Promise<T[]> {
    return await manager.getRepository(entityClass).find({
      where: { outputIndex: lastIndex + 1 } as any
    });
  }

  async getDepositTx<T extends ObjectLiteral>(
    manager: EntityManager,
    entityClass: EntityTarget<T>,
    sequence: number,
    metadata: string
  ): Promise<T | null> {
    return await manager.getRepository(entityClass).findOne({
      where: { sequence, metadata } as any
    });
  }

  public async getCoin<T extends ObjectLiteral>(
    manager: EntityManager,
    entityClass: EntityTarget<T>,
    metadata: string
  ): Promise<T | null> {
    return await manager.getRepository(entityClass).findOne({
      where: { l2Metadata: metadata } as any
    });
  }

  public async getLastOutputFromDB<T extends ObjectLiteral>(
    manager: EntityManager,
    entityClass: EntityTarget<T>
  ): Promise<T[]> {
    return await manager.getRepository<T>(entityClass).find({
      order: { outputIndex: 'DESC' } as any,
      take: 1
    });
  }

  public async getLastOutputIndex<T extends ObjectLiteral>(
    manager: EntityManager,
    entityClass: EntityTarget<T>
  ): Promise<number> {
    const lastOutput = await this.getLastOutputFromDB(manager, entityClass);
    const lastIndex = lastOutput.length == 0 ? -1 : lastOutput[0].outputIndex;
    return lastIndex;
  }

  public async getCheckpointBlockHeight<T extends ObjectLiteral>(
    manager: EntityManager,
    entityClass: EntityTarget<T>
  ): Promise<number> {
    const lastOutput = await this.getLastOutputFromDB(manager, entityClass);
    return lastOutput.length == 0 ? 0 : lastOutput[0].checkpointBlockHeight;
  }

  ///
  /// SAVE
  ///
  public async saveEntity<T extends ObjectLiteral>(
    manager: EntityManager,
    entityClass: EntityTarget<T>,
    entity: T
  ): Promise<T> {
    return await manager.getRepository(entityClass).save(entity);
  }

  ///
  ///  UTIL
  ///
  public async fetchEvents(lcd: any, height: number, eventType: string): Promise<any[]> {
    const searchRes = await lcd.tx.search({
      events: [{ key: 'tx.height', value: (height + 1).toString() }]
    });
    return searchRes.txs
      .flatMap((tx) => tx.logs ?? [])
      .flatMap((log) => log.events)
      .filter((evt) => evt.type === eventType);
  }

  public eventsToAttrMap(event: any): { [key: string]: string } {
    return event.attributes.reduce((obj, attr) => {
      obj[attr.key] = attr.value;
      return obj;
    }, {});
  }
  
  public parseData(attrMap: { [key: string]: string }): {
    [key: string]: string;
  } {
    return JSON.parse(attrMap['data']);
  }

  ///
  /// L1 HELPER
  ///

  

  ///
  /// L2 HELPER
  ///
  public calculateOutputEntity(
    lastIndex: number,
    blockInfo: BlockInfo,
    storageRoot: string,
    checkpointBlockHeight: number
  ) {
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

    const outputEntity = {
      outputIndex: lastIndex + 1,
      outputRoot,
      stateRoot,
      storageRoot,
      lastBlockHash,
      checkpointBlockHeight
    };
    return outputEntity;
  }
}

export default MonitorHelper;
