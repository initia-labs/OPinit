import * as http from 'http'
import { initApp } from './app'
import { KoaController } from 'koa-joi-controllers'
let server: http.Server

export async function initServer(
  controllers: KoaController[],
  port: number
): Promise<http.Server> {
  const app = await initApp(controllers)

  server = http.createServer(app.callback())

  server.listen(port, () => {
    console.log(`Listening on port ${port}`)
  })

  return server
}

export function finalizeServer(): void {
  server.close()
}
