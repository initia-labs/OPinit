import { OutputSubmitter } from './outputSubmitter'
import { outputLogger as logger } from '../../lib/logger'
import { once } from 'lodash'
import { initORM } from './db'

let jobs: OutputSubmitter[]

async function runBot(): Promise<void> {
  jobs = [new OutputSubmitter()]

  try {
    await Promise.all(
      jobs.map((job) => {
        job.run()
      })
    )
  } catch (err) {
    logger.info(err)
    stopOutput()
  }
}

function stopBot(): void {
  jobs.forEach((job) => job.stop())
}

export async function stopOutput(): Promise<void> {
  stopBot()

  logger.info('Finished OutputSubmitter')
  process.exit(0)
}

export async function startOutput(): Promise<void> {
  await initORM()
  await runBot()

  // attach graceful shutdown
  const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const
  signals.forEach((signal) => process.on(signal, once(stopOutput)))
}

if (require.main === module) {
  startOutput().catch(console.log)
}
