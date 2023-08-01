import { OutputSubmitter } from "./outputSubmitter";
import { logger } from "lib/logger";
import { delay } from "bluebird";
import { once } from 'lodash'

async function gracefulShutdown(): Promise<void> {  
    logger.info('Finished')
    process.exit(0)
}

async function main(): Promise<void> {
    const outputSubmitter = new OutputSubmitter();
    await outputSubmitter.init()
    
    for (;;) {
        try {
            await outputSubmitter.run();

            // attach graceful shutdown
            const signals = ['SIGHUP', 'SIGINT', 'SIGTERM'] as const
            signals.forEach((signal) => process.on(signal, once(gracefulShutdown)))
        }catch (err) {
            logger.error(`Error in outputSubmitterBot: ${err}`);
        }finally {
            await delay(3000)
        }
    }
}

if (require.main === module) {
    main().catch(console.log)
}