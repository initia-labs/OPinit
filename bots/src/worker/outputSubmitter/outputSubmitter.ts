import { BCS, Msg, MsgExecute, Wallet, MnemonicKey } from '@initia/initia.js';
import { INTERVAL_OUTPUT } from 'config';
import { ExecutorOutputEntity } from 'orm';
import { APIRequest } from 'lib/apiRequest';
import { delay } from 'bluebird';
import { outputLogger as logger } from 'lib/logger';
import { ErrorTypes } from 'lib/error';
import { GetOutputResponse } from 'service';
import { getConfig } from 'config';
import { sendTx } from 'lib/tx';

const config = getConfig();
const bcs = BCS.getInstance();

export class OutputSubmitter {
  private submitter: Wallet;
  private executor: Wallet;
  private apiRequester: APIRequest;
  private syncedHeight = 0;
  private isRunning = false;

  async init() {
    this.submitter = new Wallet(
      config.l1lcd,
      new MnemonicKey({ mnemonic: config.OUTPUT_SUBMITTER_MNEMONIC })
    );
    this.executor = new Wallet(
      config.l1lcd,
      new MnemonicKey({ mnemonic: config.EXECUTOR_MNEMONIC })
    );

    this.apiRequester = new APIRequest(config.EXECUTOR_URI);
    this.isRunning = true;
  }

  public name(): string {
    return 'output_submitter';
  }

  async getNextBlockHeight(): Promise<number> {
    const nextBlockHeight = await config.l1lcd.move.viewFunction<string>(
      '0x1',
      'op_output',
      'next_block_num',
      [],
      [
        bcs.serialize('address', this.executor.key.accAddress),
        bcs.serialize('string', config.L2ID)
      ]
    );
    return parseInt(nextBlockHeight);
  }

  async proposeL2Output(outputRoot: Buffer, l2BlockHeight: number) {
    const executeMsg: Msg = new MsgExecute(
      this.submitter.key.accAddress,
      '0x1',
      'op_output',
      'propose_l2_output',
      [],
      [
        bcs.serialize('address', this.executor.key.accAddress),
        bcs.serialize('string', config.L2ID),
        bcs.serialize('vector<u8>', outputRoot, 33), // 33 is the length of output root
        bcs.serialize('u64', l2BlockHeight)
      ]
    );
    await sendTx(this.submitter, [executeMsg]);
  }

  public async run() {
    await this.init();

    while (this.isRunning) {
      try {
        const nextBlockHeight = await this.getNextBlockHeight();
        if (nextBlockHeight <= this.syncedHeight) continue;

        const res: GetOutputResponse =
          await this.apiRequester.getQuery<GetOutputResponse>(
            `/output/height/${nextBlockHeight}`
          );
        await this.processOutputEntity(res.output, nextBlockHeight);
      } catch (err) {
        if (err.response?.data.type === ErrorTypes.NOT_FOUND_ERROR) {
          logger.warn(
            `waiting for next output. not found output from executor height`
          );
          await delay(INTERVAL_OUTPUT);
        } else {
          logger.error(err);
          this.stop();
        }
      }
    }
  }

  public async stop() {
    this.isRunning = false;
  }

  private async processOutputEntity(
    outputEntity: ExecutorOutputEntity,
    nextBlockHeight: number
  ) {
    await this.proposeL2Output(
      Buffer.from(outputEntity.outputRoot, 'hex'),
      nextBlockHeight
    );
    this.syncedHeight = nextBlockHeight;
    logger.info(
      `successfully submitted! height: ${nextBlockHeight}, output root: ${outputEntity.outputRoot}`
    );
  }
}
