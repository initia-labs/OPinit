import config from 'config'
import { RPCSocket } from 'lib/rpc'
import { Monitor } from './Monitor'
import { L1Monitor } from './L1Monitor'
import { L2Monitor } from './L2Monitor'
import { executorController } from 'controller'

import { logger } from 'lib/logger'
import { initORM, finalizeORM } from './db'
import { initServer, finalizeServer } from 'loader'
import { once } from 'lodash'
import { WalletType, initWallet } from 'lib/wallet'

let monitors: Monitor[]

export async function runBot(): Promise<void> {
  monitors = [
    new L1Monitor(config.l1lcd, new RPCSocket(config.L1_RPC_URI, 10000)),
    new L2Monitor(config.l2lcd, new RPCSocket(config.L2_RPC_URI, 10000)),
  ]

  await Promise.all(monitors.map((monitor) => {
    monitor.run()
  }))
}

export function stopBot(): void {
  monitors.forEach((monitor) => monitor.stop())
}

async function gracefulShutdown(): Promise<void> {
  stopBot()

  logger.info('Closing listening port')
  finalizeServer()

  logger.info('Closing DB connection')
  await finalizeORM()

  logger.info('Finished')
  process.exit(0)
}

async function main(): Promise<void> {
  await initORM()
  await initServer(executorController)
  initWallet(WalletType.Executor, config.l2lcd)
  await runBot()

  // attach graceful shutdown
  const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const
  signals.forEach((signal) => process.on(signal, once(gracefulShutdown)))
}

if (require.main === module) {
  main().catch(console.log)
}
