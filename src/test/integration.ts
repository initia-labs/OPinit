import { startExecutor } from 'worker/bridgeExecutor';
import Bridge, { build } from './utils/Bridge';
import { spawn } from 'child_process';
import { initORM } from 'worker/bridgeExecutor/db';
import DockerHelper from './utils/DockerHelper';
import * as path from 'path';
import {Config} from 'config';



async function main() {
  const config = Config.getConfig();
  console.log(config.L1_LCD_URI);
  console.log(config.L2ID);

  const submissionInterval = 10;
  const finalizeTime = 10;
  const bridge = new Bridge(
    submissionInterval,
    finalizeTime,
    config.L2ID,
    path.join(__dirname, 'contract')
  );
  await bridge.deployBridge();
}

if (require.main === module) {
  main();
}
