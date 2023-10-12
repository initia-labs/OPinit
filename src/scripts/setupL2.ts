import { MsgExecute, MsgPublish, BCS } from '@initia/initia.js';
import * as fs from 'fs';
import * as path from 'path';
import { sendTx } from 'lib/tx';
import { getConfig } from 'config';
import {
  build,
  executor,
  challenger,
  outputSubmitter
} from 'test/utils/helper';

const config = getConfig();
export const bcs = BCS.getInstance();

class L2Initializer {
  l2id = config.L2ID;
  moduleName = this.l2id.split('::')[1];
  contractDir = path.join(__dirname, 'contract');

  constructor(
    public submissionInterval,
    public finalizedTime,
    public l2StartBlockHeight
  ) {}
  // update module name in l2id.move
  updateL2ID() {
    const filePath = path.join(this.contractDir, 'sources', 'l2id.move');
    const fileContent = fs.readFileSync(filePath, 'utf-8');
    const updatedContent = fileContent.replace(
      /(addr::)[^\s{]+( \{)/g,
      `$1${this.moduleName}$2`
    );
    fs.writeFileSync(filePath, updatedContent, 'utf-8');
  }

  publishL2IDMsg(module: string) {
    return new MsgPublish(executor.key.accAddress, [module], 0);
  }

  bridgeInitializeMsg(l2id: string) {
    return new MsgExecute(
      executor.key.accAddress,
      '0x1',
      'op_bridge',
      'initialize',
      [l2id],
      []
    );
  }

  outputInitializeMsg(
    submissionInterval: number,
    finalizedTime: number,
    l2StartBlockHeight: number
  ) {
    return new MsgExecute(
      executor.key.accAddress,
      '0x1',
      'op_output',
      'initialize',
      [this.l2id],
      [
        bcs.serialize('u64', submissionInterval),
        bcs.serialize('address', outputSubmitter.key.accAddress),
        bcs.serialize('address', challenger.key.accAddress),
        bcs.serialize('u64', finalizedTime),
        bcs.serialize('u64', l2StartBlockHeight)
      ]
    );
  }

  bridgeRegisterTokenMsg(coinType: string) {
    return new MsgExecute(
      executor.key.accAddress,
      '0x1',
      'op_bridge',
      'register_token',
      [this.l2id, coinType],
      []
    );
  }

  async initialize() {
    this.updateL2ID();
    const module = await build(this.contractDir, this.moduleName);
    const msgs = [
      this.publishL2IDMsg(module),
      this.bridgeInitializeMsg(this.l2id),
      this.outputInitializeMsg(
        this.submissionInterval,
        this.finalizedTime,
        this.l2StartBlockHeight
      ),
      this.bridgeRegisterTokenMsg(`0x1::native_uinit::Coin`)
    ];
    await sendTx(executor, msgs);
  }
}

async function main() {
  if (!process.env.SUB_INTV) {
    throw new Error('SUB_INTV is not set');
  }
  if (!process.env.FIN_TIME) {
    throw new Error('FIN_TIME is not set');
  }
  if (!process.env.L2_HEIGHT) {
    throw new Error('L2_HEIGHT is not set');
  }

  const initializer = new L2Initializer(
    process.env.SUB_INTV, // submissionInterval
    process.env.FIN_TIME, // finalizedTime
    process.env.L2_HEIGHT // l2StartBlockHeight
  );

  console.log('=========Initializing L2=========');
  console.log('submissionInterval: ', initializer.submissionInterval);
  console.log('finalizedTime: ', initializer.finalizedTime);
  console.log('l2StartBlockHeight: ', initializer.l2StartBlockHeight);

  await initializer.initialize();
  console.log('=========L2 Initialized Done=========');
}

if (require.main === module) {
  main();
}
