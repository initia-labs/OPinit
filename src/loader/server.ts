import * as http from 'http'
import config from 'config'
import { logger } from 'lib/logger'
import { initApp } from './app'

let server: http.Server

export async function initServer(): Promise<http.Server> {
  const app = await initApp()

  server = http.createServer(app.callback())

  server.listen(config.SERVER_PORT, () => {
    logger.info(`Listening on port ${config.SERVER_PORT}`)
  })

  return server
}

export function finalizeServer(): void {
  server.close()
}
