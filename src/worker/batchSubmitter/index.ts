import { initORM, finalizeORM } from './db';
import { executorLogger as logger } from 'lib/logger';
import { BatchSubmitter } from './batchSubmitter';
import { initServer, finalizeServer } from 'loader';
import { batchController } from 'controller';
import { once } from 'lodash';
import config from 'config';

let jobs: BatchSubmitter[] = [];

async function runBot(): Promise<void> {
  jobs = [new BatchSubmitter()];

  await Promise.all(
    jobs.map((job) => {
      job.run();
    })
  );
}

function stopBot(): void {
  jobs.forEach((job) => job.stop());
}

async function gracefulShutdown(): Promise<void> {
  stopBot();

  logger.info('Closing listening port');
  finalizeServer();

  logger.info('Closing DB connection');
  await finalizeORM();

  logger.info('Finished');
  process.exit(0);
}

async function main(): Promise<void> {
  await initORM();
  await initServer(batchController, config.BATCH_PORT);

  await runBot();

  // attach graceful shutdown
  const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const;
  signals.forEach((signal) => process.on(signal, once(gracefulShutdown)));
}

if (require.main === module) {
  main().catch(console.log);
}
