import config from 'config';
import { RPCSocket } from 'lib/rpc';
import { Monitor } from './Monitor';
import { L1Monitor } from './L1Monitor';
import { L2Monitor } from './L2Monitor';
import { executorController } from 'controller';

import { executorLogger as logger } from 'lib/logger';
import { initORM, finalizeORM } from './db';
import { initServer, finalizeServer } from 'loader';
import { once } from 'lodash';
import { WalletType, initWallet } from 'lib/wallet';

let monitors: Monitor[];

export async function runBot(): Promise<void> {
  monitors = [
    new L1Monitor(new RPCSocket(config.L1_RPC_URI, 1000, logger), logger),
    new L2Monitor(new RPCSocket(config.L2_RPC_URI, 1000, logger), logger)
  ];
  try {
    await Promise.all(
      monitors.map((monitor) => {
        monitor.run();
      })
    );
  } catch (err) {
    logger.error(err);
    gracefulShutdown();
  }
}

export function stopBot(): void {
  monitors.forEach((monitor) => monitor.stop());
}

export async function gracefulShutdown(): Promise<void> {
  stopBot();

  logger.info('Closing listening port');
  finalizeServer();

  logger.info('Closing DB connection');
  await finalizeORM();

  logger.info('Finished');
  process.exit(0);
}

export async function startExecutor(): Promise<void> {
  await initORM();
  await initServer(executorController, config.EXECUTOR_PORT);
  initWallet(WalletType.Executor, config.l2lcd);
  logger.info('executor l2id :', config.L2ID);
  await runBot();

  // attach graceful shutdown
  const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const;
  signals.forEach((signal) => process.on(signal, once(gracefulShutdown)));
}

if (require.main === module) {
  startExecutor().catch(console.log);
}

export { monitors };
