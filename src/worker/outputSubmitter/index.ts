import { OutputSubmitter } from './outputSubmitter';
import { outputLogger as logger } from 'lib/logger';
import { once } from 'lodash';
import axios from 'axios';
import { getConfig } from 'config';

const config = getConfig();
let jobs: OutputSubmitter[];

async function runBot(): Promise<void> {
  const outputSubmitter = new OutputSubmitter();

  jobs = [outputSubmitter];

  try {
    await Promise.all(
      jobs.map((job) => {
        job.run();
      })
    );
  } catch (err) {
    logger.error(err);
    stopOutput();
  }
}

function stopBot(): void {
  jobs.forEach((job) => job.stop());
}

export async function stopOutput(): Promise<void> {
  stopBot();

  logger.info('Finished OutputSubmitter');
  process.exit(0);
}

export async function startOutput(): Promise<void> {
  await checkExecutor();
  await runBot();

  // attach graceful shutdown
  const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const;
  signals.forEach((signal) => process.on(signal, once(stopOutput)));
}

export const checkExecutor = async (timeout = 60_000) => {
  const startTime = Date.now();

  while (Date.now() - startTime < timeout) {
    try {
      const response = await axios.get(config.EXECUTOR_URI + '/health');
      if (response.status === 200) return;
    } catch {
      logger.info('waiting for executor');
      await new Promise((res) => setTimeout(res, 5_000));
      continue;
    }
  }
};

if (require.main === module) {
  startOutput().catch(console.log);
}
