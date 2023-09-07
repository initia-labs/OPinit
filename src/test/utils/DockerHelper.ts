import { upAll } from 'docker-compose';
import { exec as execCb } from 'child_process';
import { promisify } from 'util';

export default class DockerHelper {
  constructor(public path: string) {}

  async start() {
    console.log('Starting docker containers...');
    const result = await upAll({ cwd: this.path, log: false });
    return result;
  }

  async stopDocker(scriptPath: string): Promise<void> {
    const exec = promisify(execCb);
    try {
      const { stdout, stderr } = await exec(
        `sh ${scriptPath}/docker-compose-reset`
      );

      if (stderr) {
        console.warn(`stderr: ${stderr}`);
      }
    } catch (error) {
      console.warn(`Error: ${error.message}`);
    }
  }

  async stop() {
    console.log('Stopping docker containers...');
    await this.stopDocker(this.path);
    console.log('Successfully stopped docker containers');
  }
}
