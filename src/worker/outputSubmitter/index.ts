import { OutputSubmitter } from "./outputSubmitter";
import { initORM, finalizeORM } from "./db";
import { logger } from "lib/logger";
import Bluebird from "bluebird";
import { once } from 'lodash'

async function gracefulShutdown(): Promise<void> {  
    logger.info('Closing DB connection')
    await finalizeORM()
  
    logger.info('Finished')
    process.exit(0)
}

async function main(): Promise<void> {
    const outputSubmitter = new OutputSubmitter();
    await outputSubmitter.init()
    await initORM()
    
    for (;;) {
        try {
            await outputSubmitter.run();

            // attach graceful shutdown
            const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const
            signals.forEach((signal) => process.on(signal, once(gracefulShutdown)))
        }catch (err) {
            logger.error(`Error in outputSubmitterBot: ${err}`);
        }finally {
            await Bluebird.Promise.delay(3000)
        }
    }
}

if (require.main === module) {
    main().catch(console.log)
}