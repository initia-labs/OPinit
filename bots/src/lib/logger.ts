import * as winston from 'winston';
import * as DailyRotateFile from 'winston-daily-rotate-file';
import { config } from 'config';

function getLogsSubdir() {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(
    2,
    '0'
  )}-${String(d.getDate()).padStart(2, '0')}`;
}

function createLogger(name: string) {
  const formats = [winston.format.errors({ stack: true })];

  if (!config.USE_LOG_FILE) {
    formats.push(winston.format.colorize());
  }

  formats.push(
    winston.format.timestamp(),
    winston.format.printf((info) => {
      return `${info.timestamp} [${info.level} - ${name}]: ${
        info.stack || info.message
      }`;
    })
  );

  const logger = winston.createLogger({
    format: winston.format.combine(...formats),
    defaultMeta: { service: 'user-service' }
  });

  if (config.USE_LOG_FILE) {
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
  } else {
    logger.add(new winston.transports.Console());
  }

  return logger;
}

export const executorLogger = createLogger('Executor');
export const outputLogger = createLogger('Output');
export const batchLogger = createLogger('Batch');
export const challengerLogger = createLogger('Challenger');
