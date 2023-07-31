import { initORM } from './db'
import { logger } from "../../lib/logger"
import BlueBird  from 'bluebird';
import L2Monitoring from '../l2Monitoring';
import { BatchSubmitter } from './batchSubmitter';
import { WalletType, initWallet } from 'lib/wallet';
import config from 'config';

(async () => {
    const batchSubmitter = new BatchSubmitter();
    const l2Monitoring = new L2Monitoring();
    await batchSubmitter.init()
    await l2Monitoring.init()
    await initORM()
    initWallet(WalletType.BatchSubmitter, config.l1lcd)

    for (;;) {
        try {
            if (!await l2Monitoring.isValid()) {
                logger.info(`[BatchSubmitter] L2 is invalid. stop batch submitter.`);
            } else {
                await batchSubmitter.run();            
            }
        }catch (err) {
            logger.error(`Error in batchBot: ${err}`);
        }finally {
            await BlueBird.Promise.delay(3000)
        }
    }
})();