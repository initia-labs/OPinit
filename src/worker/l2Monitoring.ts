import { logger } from "../lib/logger"
import config from "../config"
import { lookupInvalidBlock} from "../lib/rpc";

class L2Monitoring{
    private invalidBlockHeight: number;
    private l2Url: string;

    constructor(){
        this.invalidBlockHeight = 0
        this.l2Url = config.L2_RPC_URI
    }

    async init() {
        this.invalidBlockHeight = 0
        this.l2Url = config.L2_RPC_URI
    }

    public async isValid(): Promise<boolean> {
        try{
            const invalidBlock = await lookupInvalidBlock(this.l2Url);

            if (invalidBlock) {
                logger.info(`[L2Monitoring] invalid ${invalidBlock.height} block found ${invalidBlock.reason}`);
                this.invalidBlockHeight = parseInt(invalidBlock.height)
                return false
            }
            this.invalidBlockHeight = 0
            return true
        } catch (err) {
            throw new Error(`Error in L2Monitoring: ${err}`);
        }
    }

    public getInvalidBlockHeight(): number {
        return this.invalidBlockHeight
    }

}

export default L2Monitoring;