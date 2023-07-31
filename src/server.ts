import * as http from 'http'
import * as Bluebird from 'bluebird'
import * as sentry from '@sentry/node'

import { getDB, initORM } from './orm'
import config from './config'
import createApp from './createApp'
import { logger } from './lib/logger'


const packageJson = require('../package.json')

Bluebird.config({
  longStackTraces: true
})

global.Promise = Bluebird as any

process.on('unhandledRejection', (err) => {
  sentry.captureException(err)
  throw err
})

export async function createServer() {
  await initORM()
  const [dbSource] = getDB()

  const app = await createApp()
  const server = http.createServer(app.callback())

  server.listen(config.SERVER_PORT, () => {
    logger.info(`${packageJson.description} is listening on port ${config.SERVER_PORT}`)
  })

  return dbSource
}

if (require.main === module) {
  createServer().catch((err) => {
    logger.error(err)
  })
}