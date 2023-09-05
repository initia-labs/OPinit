import { startExecutor } from 'worker/bridgeExecutor';
import Bridge, { build } from './utils/Bridge';
import { spawn } from 'child_process';
import { initORM } from 'worker/bridgeExecutor/db';
import * as crypto from 'crypto';
import DockerHelper from './utils/DockerHelper';
import * as path from 'path';

const runExecutor = (env: NodeJS.ProcessEnv) => {
  // const env = Object.assign({}, process.env, {
  //     L2ID: l2id,
  //     EXECUTOR_URI: 'http://localhost:3000',
  // });
  const child = spawn('npm', ['run', 'start'], { env });
  child.unref();
};

async function main() {
  const l2id = '0x56ccf33c45b99546cd1da172cf6849395bbf8573::s10tt4::Minitia';
  const submissionInterval = 10;
  const finalizeTime = 10;
  const bridge = new Bridge(submissionInterval, finalizeTime, l2id, 'contract');

  const docker = new DockerHelper();
  const imageName = 'initia:latest';
  const contextPath = path.join(__dirname, 'docker/initia');

  const buildRes = await docker.buildImage(
    {
      context: contextPath,
      src: ['Dockerfile']
    },
    {
      t: imageName
    }
  );
  console.log(buildRes);
  // await docker.runImage(imageName)
  // await docker.stopContainer(imageName)
  // await docker.removeImage(imageName)
}

if (require.main === module) {
  main();
}
