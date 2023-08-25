import { RPCSocket } from 'lib/rpc';
import { L1Monitor } from './L1Monitor';
import { Monitor } from './Monitor';
import { Challenger } from './challenger';
import { initORM, finalizeORM } from './db';
import { challengerLogger as logger } from 'lib/logger';
import { once } from 'lodash';
import config from 'config';
import { L2Monitor } from './L2Monitor';

let monitors: (Monitor | Challenger)[];

async function runBot(): Promise<void> {
  const challenger = new Challenger();

  // use to sync with bridge latest state
  await challenger.fetchBridgeState();

  monitors = [
    new L1Monitor(new RPCSocket(config.L1_RPC_URI, 10000, logger)),
    new L2Monitor(new RPCSocket(config.L2_RPC_URI, 10000, logger)),
    challenger
  ];

  await Promise.all(
    monitors.map((monitor) => {
      monitor.run();
    })
  );
}

function stopBot(): void {
  monitors.forEach((monitor) => monitor.stop());
}

async function gracefulShutdown(): Promise<void> {
  stopBot();

  logger.info('Closing DB connection');
  await finalizeORM();

  logger.info('Finished');
  process.exit(0);
}

async function main(): Promise<void> {
  await initORM();
  await runBot();

  const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const;
  signals.forEach((signal) => process.on(signal, once(gracefulShutdown)));
}

if (require.main === module) {
  main().catch(console.log);
}
