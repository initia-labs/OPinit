import { Challenger } from './challenger';
import { initORM, finalizeORM } from './db';
import { logger } from 'lib/logger';
import { once } from 'lodash';

async function gracefulShutdown(): Promise<void> {
  logger.info('Closing DB connection');
  await finalizeORM();

  logger.info('Finished');
  process.exit(0);
}

async function main(): Promise<void> {
  await initORM();

  const challenger = new Challenger();
  await challenger.init();

  const jobs = [
    challenger.monitorL1Deposit(),
    challenger.monitorL2Withdrawal()
  ];
  await Promise.all(jobs);

  const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const;
  signals.forEach((signal) => process.on(signal, once(gracefulShutdown)));
}

if (require.main === module) {
  main().catch(console.log);
}
