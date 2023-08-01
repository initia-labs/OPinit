import { OutputEntity } from 'orm'
import { getDB } from 'worker/bridgeExecutor/db'
import { APIError, ErrorTypes } from "lib/error";

export async function getOutput(outputIndex: number): Promise<OutputEntity> {
  const [db] = getDB()
  const queryRunner = db.createQueryRunner('slave')

  try {
    const qb = queryRunner.manager
      .createQueryBuilder(OutputEntity, 'output')
      .where('output.outputIndex = :outputIndex', { outputIndex })

    const output = await qb.getOne()

    if (!output) {
      throw new APIError(ErrorTypes.NOT_FOUND_ERROR)
    }

    return output
  } finally {
    queryRunner.release()
  }
}