import { initORM, finalizeORM } from './db'
import { logger } from "../../lib/logger"
import BlueBird  from 'bluebird';
import L2Monitoring from '../l2Monitoring';
import { BatchSubmitter } from './batchSubmitter';
import { initServer, finalizeServer } from 'loader';
import { batchController } from 'controller';
import { once } from 'lodash'

async function gracefulShutdown(): Promise<void> {
    logger.info('Closing listening port')
    finalizeServer()
  
    logger.info('Closing DB connection')
    await finalizeORM()
  
    logger.info('Finished')
    process.exit(0)
}


async function main(): Promise<void> {
    await initORM()
    await initServer(batchController, 3001)

    const batchSubmitter = new BatchSubmitter();
    const l2Monitoring = new L2Monitoring();
    await batchSubmitter.init()
    await l2Monitoring.init()

    // initWallet(WalletType.BatchSubmitter, config.l1lcd)

    for (;;) {
        try {
            if (!await l2Monitoring.isValid()) {
                logger.info(`[BatchSubmitter] L2 is invalid. stop batch submitter.`);
            } else {
                await batchSubmitter.run();            
            }
            // attach graceful shutdown
            const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const
            signals.forEach((signal) => process.on(signal, once(gracefulShutdown)))
        }catch (err) {
            logger.error(`Error in batchBot: ${err}`);
        }finally {
            await BlueBird.Promise.delay(3000)
        }
    }

}

if (require.main === module) {
    main().catch(console.log)
}
  
