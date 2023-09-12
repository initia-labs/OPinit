import Bridge from './utils/Bridge';
import DockerHelper from './utils/DockerHelper';
import * as path from 'path';
import { startBatch } from 'worker/batchSubmitter';
import { startOutput } from 'worker/outputSubmitter';
import { startExecutor } from 'worker/bridgeExecutor';
import { startChallenger } from 'worker/challenger';
import { Config } from 'config';
import { delay } from 'bluebird';
import { TxBot } from './utils/TxBot';
import {
  Wallet as mWallet,
  MnemonicKey as mMnemonicKey
} from '@initia/minitia.js';

import {
  Wallet as iWallet,
  MnemonicKey as iMnemonicKey
} from '@initia/initia.js';
import { checkExecutor } from './utils/helper';

const config = Config.getConfig();
const docker = new DockerHelper(path.join(__dirname, '..', '..'));

async function main() {
  try {
    await docker.start();
    await delay(20_000); // time for setting up docker

    await setupBridge(10, 10, 10);
    await startBot();
    await startTxBot();
  } catch (err) {
    console.log(err);
  }
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
  await bridge.deployBridge();
}

async function startBot() {
  startBatch();
  startExecutor();
  startChallenger(false);
  
  await checkExecutor(); // WARN: run after executor started
  startOutput();
}

async function startTxBot() {
  const txBot = new TxBot();

  const l1sender = txBot.createWallet(
    config.l1lcd,
    iWallet,
    iMnemonicKey,
    'banner december bunker moral nasty glide slow property pen twist doctor exclude novel top material flee appear imitate cat state domain consider then age'
  );
  const l2sender = txBot.createWallet(
    config.l2lcd,
    iWallet,
    iMnemonicKey,
    'banner december bunker moral nasty glide slow property pen twist doctor exclude novel top material flee appear imitate cat state domain consider then age'
  );
  const l1receiver = txBot.createWallet(
    config.l1lcd,
    mWallet,
    mMnemonicKey,
    'diamond donkey opinion claw cool harbor tree elegant outer mother claw amount message result leave tank plunge harbor garment purity arrest humble figure endless'
  );
  const l2receiver = txBot.createWallet(
    config.l2lcd,
    mWallet,
    mMnemonicKey,
    'diamond donkey opinion claw cool harbor tree elegant outer mother claw amount message result leave tank plunge harbor garment purity arrest humble figure endless'
  );

  // create account number
  await txBot.sendCoin(l1sender, l1receiver, 1_000_000, 'uinit');
  await txBot.sendCoin(l2sender, l2receiver, 1_000_000, 'umin');

  try {
    await txBot.deposit(l1sender, l2receiver, 1_000);
    // await txBot.withdrawal(l2receiver, 100); // WARN: run after deposit done
  } catch (err) {
    console.log(err);
  }
}

if (require.main === module) {
  main();
}
