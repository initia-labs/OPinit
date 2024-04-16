import { RPCClient, RPCSocket } from '../../lib/rpc'
import { Monitor } from '../../lib/monitor'
import { Challenger } from './challenger'
import { initORM, finalizeORM } from './db'
import { challengerLogger as logger } from '../../lib/logger'
import { once } from 'lodash'
import { L1Monitor } from './monitor_l1'
import { L2Monitor } from './monitor_l2'
import { config } from '../../config'

let monitors: (Monitor | Challenger)[]

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
    new Challenger(logger)
  ]
  try {
    await Promise.all(
      monitors.map((monitor) => {
        monitor.run()
      })
    )
  } catch (err) {
    logger.info(err)
    stopChallenger()
  }
}

function stopBot(): void {
  monitors.forEach((monitor) => monitor.stop())
}

export async function stopChallenger(): Promise<void> {
  stopBot()

  logger.info('Closing DB connection')
  await finalizeORM()

  logger.info('Finished Challenger')
  process.exit(0)
}

export async function startChallenger(): Promise<void> {
  await initORM()
  await runBot()

  const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const
  signals.forEach((signal) => process.on(signal, once(stopChallenger)))
}

if (require.main === module) {
  startChallenger().catch(console.log)
}
