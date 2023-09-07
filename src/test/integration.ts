import Bridge from './utils/Bridge';
import DockerHelper from './utils/DockerHelper';
import * as path from 'path';
import { startBatch } from 'worker/batchSubmitter';
import { startOutput } from 'worker/outputSubmitter';
import { startExecutor } from 'worker/bridgeExecutor';
import { startChallenger } from 'worker/challenger';
import { Config } from 'config';
import { delay } from 'bluebird';

const config = Config.getConfig();
const docker = new DockerHelper(path.join(__dirname, '..', '..'));

async function main() {
  try {
    await setupBridge();
    await startBot();
  } catch (err) {
    console.log(err);
  }
}

async function setupBridge() {
  await docker.start();
  await delay(30_000);

  const bridge = new Bridge(
    10,
    10,
    1,
    config.L2ID,
    path.join(__dirname, 'contract')
  );
  await bridge.deployBridge();
}

async function startBot() {
  startExecutor();
  await delay(10_000);
  startOutput();
  startBatch();
  startChallenger();
}

if (require.main === module) {
  main();
}
