import config from '../config';
import * as winston from 'winston';
import * as DailyRotateFile from 'winston-daily-rotate-file';
import chalk from 'chalk';

function pad(input: number | string, width: number, z = '0') {
  const n = typeof input === 'number' ? input.toString() : input;
  return n.padStart(width, z);
}

function getDateString() {
  const d = new Date();
  return `${pad(d.getMonth() + 1, 2)}-${pad(d.getDate(), 2)} ${pad(
    d.getHours(),
    2
  )}:${pad(d.getMinutes(), 2)}`;
}

function getLogsSubdir() {
  const d = new Date();
  return `${d.getFullYear()}-${pad(d.getMonth() + 1, 2)}-${pad(
    d.getDate(),
    2
  )}`;
}

const print = winston.format.printf((info) => {
  let level;

  if (!config.USE_LOG_FILE) {
    // Do not colorize when writing to file
    if (info.level === 'error') {
      level = chalk.red(info.level.toUpperCase());
    } else if (info.level === 'warn') {
      level = chalk.yellow(info.level.toUpperCase());
    } else {
      level = chalk.green(info.level.toUpperCase());
    }
  }

  const log = `${getDateString()} [${level ? level : info.level}]: ${
    info.message
  }`;

  return log;
});

function createLogger(name: string) {
  const logger = winston.createLogger({
    level: 'info',
    format: winston.format.combine(
      winston.format.errors({ stack: true }),
      print
    ),
    defaultMeta: { service: 'user-service' }
  });

  if (config.USE_LOG_FILE) {
    //
    // - Write to all logs with level `info` and below to `combined.log`
    // - Write all logs error (and below) to `error.log`.
    //
    logger.add(
      new DailyRotateFile({
        level: 'error',
        filename: `logs/${getLogsSubdir()}/${name}_error.log`,
        zippedArchive: true
      })
    );

    logger.add(
      new DailyRotateFile({
        filename: `logs/${getLogsSubdir()}/${name}_combined.log`,
        zippedArchive: true
      })
    );
  }

  if (!config.USE_LOG_FILE || process.env.HOST_ENV === 'docker') {
    logger.add(new winston.transports.Console());
  }

  return logger;
}

export const executorLogger = createLogger('executor');
export const batchLogger = createLogger('batch');
export const challengerLogger = createLogger('challenger');
export const outputLogger = createLogger('output');
