import { RPCSocket } from 'lib/rpc';
import { L1Monitor } from './L1Monitor';
import { Monitor } from 'worker/bridgeExecutor/Monitor';
import { Challenger } from './challenger';
import { initORM, finalizeORM } from './db';
import { challengerLogger as logger } from 'lib/logger';
import { once } from 'lodash';
import { L2Monitor } from './L2Monitor';
import { getConfig } from 'config';

const config = getConfig();

let monitors: (Monitor | Challenger)[];

async function runBot(isFetch?: boolean): Promise<void> {
  const challenger = new Challenger();

  // use to sync with bridge latest state
  if (isFetch) await challenger.fetchBridgeState();

  monitors = [
    new L1Monitor(new RPCSocket(config.L1_RPC_URI, 10000, logger), logger),
    new L2Monitor(new RPCSocket(config.L2_RPC_URI, 10000, logger), logger),
    challenger
  ];
  try {
    await Promise.all(
      monitors.map((monitor) => {
        monitor.run();
      })
    );
  } catch (err) {
    logger.error(err);
    stopChallenger();
  }
}

function stopBot(): void {
  monitors.forEach((monitor) => monitor.stop());
}

export async function stopChallenger(): Promise<void> {
  stopBot();

  logger.info('Closing DB connection');
  await finalizeORM();

  logger.info('Finished Challenger');
  process.exit(0);
}

export async function startChallenger(isFetch = true): Promise<void> {
  await initORM();
  await runBot(isFetch);

  const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const;
  signals.forEach((signal) => process.on(signal, once(stopChallenger)));
}

if (require.main === module) {
  startChallenger().catch(console.log);
}
