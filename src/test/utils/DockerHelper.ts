import * as Dockerode from 'dockerode';
export default class Docker {
  docker: Dockerode;

  constructor() {
    this.docker = new Dockerode();
  }

  async buildImage(
    file: string | NodeJS.ReadableStream | Dockerode.ImageBuildContext,
    options: Dockerode.ImageBuildOptions
  ) {
    const stream: NodeJS.ReadableStream = await this.docker.buildImage(
      file,
      options
    );
    return await new Promise((resolve, reject) => {
      this.docker.modem.followProgress(stream, (err, res) =>
        err ? reject(err) : resolve(res)
      );
    });
  }

  async runImage(imageName: string) {
    // //promise
    // this.docker.run(imageName, ['bash', '-c', 'uname -a'], process.stdout).then(function(data) {
    //   const output = data[0];
    //   const container = data[1];
    //   console.log(output.StatusCode);
    //   return container.remove();
    // }).then(function(data) {
    //   console.log('container removed');
    // }).catch(function(err) {
    //   console.log(err);
    // });
    try {
      const container = await this.docker.createContainer({
        Image: imageName,
        AttachStdin: false,
        AttachStdout: true,
        AttachStderr: true,
        Tty: true
      });

      await container.start();

      return new Promise<void>((resolve, reject) => {
        container
          .logs({
            follow: true,
            stdout: true,
            stderr: true
          })
          .then((stream) => {
            stream.on('data', (data: any) => console.log(data.toString()));
            stream.on('end', () => {
              console.log('Container logs stream ended.');
              resolve();
            });
            stream.on('error', (error) => {
              console.error('Error streaming container logs:', error);
              reject(error);
            });
          })
          .catch((error) => {
            console.error('Error fetching container logs:', error);
            reject(error);
          });
      });
    } catch (error) {
      console.error(`Error running image ${imageName}:`, error);
      throw error;
    }
  }

  async stopContainer(imageName: string) {
    try {
      const containers = await this.docker.listContainers({ all: true });
      const relevantContainers = containers.filter(
        (containerInfo) => containerInfo.Image === imageName
      );

      for (const containerInfo of relevantContainers) {
        const container = this.docker.getContainer(containerInfo.Id);
        await container.stop();
        console.log(`Stopped container with ID: ${containerInfo.Id}`);
      }
    } catch (error) {
      console.error(
        `Error stopping containers using image ${imageName}:`,
        error
      );
    }
  }

  async removeImage(imageName: string) {
    try {
      const containers = await this.docker.listContainers({ all: true });
      const relevantContainers = containers.filter(
        (containerInfo) => containerInfo.Image === imageName
      );

      // Stop all containers using the image before removal
      for (const containerInfo of relevantContainers) {
        await this.stopContainer(containerInfo.Id);
      }

      await this.docker.getImage(imageName).remove();
      console.log(`Image ${imageName} removed.`);
    } catch (error) {
      console.error(`Error removing image ${imageName}:`, error);
      throw error;
    }
  }
}
