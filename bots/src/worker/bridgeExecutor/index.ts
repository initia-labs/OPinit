import { RPCClient, RPCSocket } from '../../lib/rpc'
import { L1Monitor } from '../../lib/monitor/l1'
import { L2Monitor } from '../../lib/monitor/l2'
import { executorController } from '../../controller'

import { executorLogger as logger } from '../../lib/logger'
import { initORM, finalizeORM } from './db'
import { initServer, finalizeServer } from '../../loader'
import { once } from 'lodash'
import { config } from '../../config'
import { Resurrector } from './Resurrector'

let monitors

async function runBot(): Promise<void> {
  monitors = [
    new L1Monitor(
      new RPCSocket(config.L1_RPC_URI, 10000, logger),
      new RPCClient(config.L1_RPC_URI, logger),
      logger
    ),
    new L2Monitor(
      new RPCSocket(config.L2_RPC_URI, 10000, logger),
      new RPCClient(config.L2_RPC_URI, logger),
      logger
    ),
    new Resurrector(logger)
  ]
  try {
    await Promise.all(
      monitors.map((monitor) => {
        monitor.run()
      })
    )
  } catch (err) {
    logger.info(err)
    stopExecutor()
  }
}

function stopBot(): void {
  monitors.forEach((monitor) => monitor.stop())
}

export async function stopExecutor(): Promise<void> {
  stopBot()

  logger.info('Closing listening port')
  finalizeServer()

  logger.info('Closing DB connection')
  await finalizeORM()

  logger.info('Finished Executor')
  process.exit(0)
}

export async function startExecutor(): Promise<void> {
  try {
    await initORM()

    await initServer(executorController, config.EXECUTOR_PORT)

    if (!config.ENABLE_API_ONLY) {
      await runBot()
    }
  } catch (err) {
    throw new Error(err)
  }

  // attach graceful shutdown
  const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const
  signals.forEach((signal) => process.on(signal, once(stopExecutor)))
}

if (require.main === module) {
  startExecutor().catch(console.log)
}

export { monitors }
