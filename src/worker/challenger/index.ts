import { Challenger } from "./challenger";
import { initORM } from "./db";
import { logger } from "lib/logger";
import Bluebird from "bluebird";

(async () => {
    const challenger = new Challenger();
    await challenger.init()
    await initORM()
    
    for (;;) {
        try {
            await challenger.run();
        }catch (err) {
            logger.error(`Error in challegnerBot: ${err}`);
        }finally {
            await Bluebird.Promise.delay(3000)
        }
    }
})();