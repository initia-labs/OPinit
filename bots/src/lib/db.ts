import 'reflect-metadata'
import Bluebird from 'bluebird'
import {
  ConnectionOptionsReader,
  DataSource,
  DataSourceOptions
} from 'typeorm'
import { PostgresConnectionOptions } from 'typeorm/driver/postgres/PostgresConnectionOptions'
import CamelToSnakeNamingStrategy from '../orm/CamelToSnakeNamingStrategy'

const debug = require('debug')('orm')

import { RecordEntity, ExecutorOutputEntity } from '../orm'

const staticOptions = {
  supportBigNumbers: true,
  bigNumberStrings: true,
  entities: [RecordEntity, ExecutorOutputEntity]
}

let DB: DataSource[] = []

function initConnection(options: DataSourceOptions): Promise<DataSource> {
  const pgOpts = options as PostgresConnectionOptions
  debug(
    `creating connection default to ${pgOpts.username}@${pgOpts.host}:${
      pgOpts.port || 5432
    }`
  )

  return new DataSource({
    ...options,
    ...staticOptions,
    namingStrategy: new CamelToSnakeNamingStrategy()
  }).initialize()
}

export async function initORM(): Promise<void> {
  const reader = new ConnectionOptionsReader()
  const options = (await reader.all()) as PostgresConnectionOptions[]

  if (options.length && !options.filter((o) => o.name === 'default').length) {
    options[0]['name' as any] = 'default'
  }

  DB = await Bluebird.map(options, (opt) => initConnection(opt))
}

export function getDB(): DataSource[] {
  if (!DB) {
    throw new Error('DB not initialized')
  }
  return DB
}

export async function finalizeORM(): Promise<void> {
  await Promise.all(DB.map((c) => c.destroy()))
}
