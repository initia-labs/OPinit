import { EntityManager, EntityTarget, ObjectLiteral } from 'typeorm';

export const ENOT_EQUAL_TX = -1;

class ChallengerHelper {
  public async getUncheckedTx<T extends ObjectLiteral>(
    manager: EntityManager,
    entityClass: EntityTarget<T>
  ): Promise<T | null> {
    const uncheckedWithdrawalTx = await manager
      .getRepository<T>(entityClass)
      .find({
        where: { isChecked: false } as any,
        order: { sequence: 'ASC' } as any,
        take: 1
      });

    if (uncheckedWithdrawalTx.length === 0) return null;
    return uncheckedWithdrawalTx[0];
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

  public async finalizeUncheckedTx<T extends ObjectLiteral>(
    manager: EntityManager,
    entityClass: EntityTarget<T>,
    entity: T
  ): Promise<void> {
    await manager.getRepository(entityClass).update(
      {
        metadata: entity.metadata,
        sequence: entity.sequence
      },
      { isChecked: true } as any
    );
  }
}

export default ChallengerHelper;
