import { MsgCreateBridge, BridgeConfig, Duration } from '@initia/initia.js';
import { sendTx } from 'lib/tx';
import { config } from 'config';
import { executor, challenger, outputSubmitter } from 'test/utils/helper';

class L2Initializer {
  bridgeId = config.BRIDGE_ID;

  constructor(
    public submissionInterval: number,
    public finalizedTime: number,
    public metadata: string
  ) {}

  MsgCreateBridge(submissionInterval: number, finalizedTime: number) {
    const bridgeConfig = new BridgeConfig(
      challenger.key.accAddress,
      outputSubmitter.key.accAddress,
      Duration.fromString(submissionInterval.toString()),
      Duration.fromString(finalizedTime.toString()),
      new Date(),
      this.metadata
    );
    return new MsgCreateBridge(executor.key.accAddress, bridgeConfig);
  }

  async initialize() {
    const msgs = [
      this.MsgCreateBridge(this.submissionInterval, this.finalizedTime)
    ];

    await sendTx(executor, msgs);
  }
}

async function main() {
  try {
    const initializer = new L2Initializer(
      config.SUBMISSION_INTERVAL,
      config.FINALIZATION_PERIOD,
      config.IBC_METADATA
    );
    console.log('=========Initializing L2=========');
    console.log('submissionInterval: ', initializer.submissionInterval);
    console.log('finalizedTime: ', initializer.finalizedTime);
    console.log('metadata: ', initializer.metadata);
    console.log('bridgeId: ', initializer.bridgeId);
    await initializer.initialize();
    console.log('=========L2 Initialized Done=========');
  } catch (e) {
    console.error(e);
  }
}

if (require.main === module) {
  main();
}
