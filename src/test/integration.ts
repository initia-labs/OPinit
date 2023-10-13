import Bridge from './utils/Bridge';
import DockerHelper from './utils/DockerHelper';
import * as path from 'path';
import { startBatch } from 'worker/batchSubmitter';
import { startOutput } from 'worker/outputSubmitter';
import { startExecutor } from 'worker/bridgeExecutor';
import { startChallenger } from 'worker/challenger';
import { Config } from 'config';
import { TxBot } from './utils/TxBot';
import { computeCoinMetadata, normalizeMetadata } from 'lib/lcd';
import { checkHealth } from './utils/helper';

const config = Config.getConfig();
const docker = new DockerHelper(path.join(__dirname, '..', '..'));

async function setup() {
  await docker.start();
  await checkHealth(config.L1_LCD_URI, 20_000)
  await checkHealth(config.L2_LCD_URI, 20_000)
  await setupBridge(10, 10, 1);
}

async function setupBridge(
  submissionInterval: number,
  finalizedTime: number,
  l2StartBlockHeight: number
) {
  const bridge = new Bridge(
    submissionInterval,
    finalizedTime,
    l2StartBlockHeight,
    config.L2ID,
    path.join(__dirname, 'contract')
  );
  const UINIT_METADATA = normalizeMetadata(computeCoinMetadata('0x1', 'uinit')); // '0x8e4733bdabcf7d4afc3d14f0dd46c9bf52fb0fce9e4b996c939e195b8bc891d9'

  await bridge.deployBridge(UINIT_METADATA);
  console.log('Bridge deployed');
}

async function startBot() {
  try {
    await Promise.all([
      startBatch(),
      startExecutor(),
      startChallenger(false), // false for not fetching executor state
      startOutput()
    ]);
  } catch (err) {
    console.log(err);
  }
}

async function startTxBot() {
  const txBot = new TxBot();

  try {
    
    await txBot.deposit(txBot.l1sender, txBot.l2receiver, 1_000);
    // await txBot.withdrawal(txBot.l2receiver, 100);          // WARN: run after deposit done
    // await txBot.claim(txBot.l1receiver, 1, 13); // WARN: run after withdrawal done
    console.log('tx bot done')
  } catch (err) {
    console.log(err);
  }
}

async function main() {
  try {
    await setup();
    await startBot();
    await startTxBot();
  } catch (err) {
    console.log(err);
  }
}

if (require.main === module) {
  main();
}
