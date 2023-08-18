import { OutputSubmitter } from './outputSubmitter';
import { logger } from 'lib/logger';
import { once } from 'lodash';

let jobs: OutputSubmitter[];

async function runBot(): Promise<void> {
  const outputSubmitter = new OutputSubmitter();

  jobs = [outputSubmitter];

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

  logger.info('Finished');
  process.exit(0);
}

async function main(): Promise<void> {
  await runBot();

  // attach graceful shutdown
  const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const;
  signals.forEach((signal) => process.on(signal, once(gracefulShutdown)));
}

if (require.main === module) {
  main().catch(console.log);
}
