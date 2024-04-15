import { initORM, finalizeORM } from '../../lib/db'
import { executorLogger as logger } from '../../lib/logger'
import { BatchSubmitter } from './batchSubmitter'
import { initServer, finalizeServer } from '../../loader'
import { batchController } from '../../controller'
import { once } from 'lodash'
import { config } from '../../config'

let jobs: BatchSubmitter[] = []

async function runBot(): Promise<void> {
  jobs = [new BatchSubmitter()]

  try {
    await Promise.all(
      jobs.map((job) => {
        job.run()
      })
    )
  } catch (err) {
    logger.info(err)
    stopBatch()
  }
}

function stopBot(): void {
  jobs.forEach((job) => job.stop())
}

export async function stopBatch(): Promise<void> {
  stopBot()

  logger.info('Closing listening port')
  finalizeServer()

  logger.info('Closing DB connection')
  await finalizeORM()

  logger.info('Finished Batch')
  process.exit(0)
}

export async function startBatch(): Promise<void> {
  await initORM()

  await initServer(batchController, config.BATCH_PORT)

  if (!config.ENABLE_API_ONLY) {
    await runBot()
  }

  // attach graceful shutdown
  const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const
  signals.forEach((signal) => process.on(signal, once(stopBatch)))
}

if (require.main === module) {
  startBatch().catch(console.log)
}
