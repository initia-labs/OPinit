import * as http from 'http'
import config from 'config'
import { logger } from 'lib/logger'
import { initApp } from './app'
import { KoaController } from 'koa-joi-controllers'
let server: http.Server

export async function initServer(controllers: KoaController[]): Promise<http.Server> {
  const app = await initApp(controllers)

  server = http.createServer(app.callback())

  server.listen(config.SERVER_PORT, () => {
    logger.info(`Listening on port ${config.SERVER_PORT}`)
  })

  return server
}

export function finalizeServer(): void {
  server.close()
}
