import {
  MsgCreateBridge,
  BridgeConfig,
  BatchInfo,
  Duration,
  Wallet,
  MnemonicKey,
  BridgeInfo,
  MsgSetBridgeInfo
} from '@initia/initia.js'
import { sendTx } from '../lib/tx'
import { config } from '../config'

export const executor = new Wallet(
  config.l1lcd,
  new MnemonicKey({ mnemonic: config.EXECUTOR_MNEMONIC })
)
export const executorL2 = new Wallet(
  config.l2lcd,
  new MnemonicKey({ mnemonic: config.EXECUTOR_MNEMONIC })
)
export const challenger = new Wallet(
  config.l1lcd,
  new MnemonicKey({ mnemonic: config.CHALLENGER_MNEMONIC })
)
export const outputSubmitter = new Wallet(
  config.l1lcd,
  new MnemonicKey({ mnemonic: config.OUTPUT_SUBMITTER_MNEMONIC })
)
export const batchSubmitter = new MnemonicKey({
  mnemonic: config.BATCH_SUBMITTER_MNEMONIC
})

class L2Initializer {
  bridgeId = config.BRIDGE_ID

  constructor(
    public submissionInterval: number,
    public finalizedTime: number,
    public metadata: string
  ) {}

  MsgCreateBridge(submissionInterval: number, finalizedTime: number) {
    const bridgeConfig = new BridgeConfig(
      challenger.key.accAddress,
      outputSubmitter.key.accAddress,
      new BatchInfo(batchSubmitter.accAddress, config.PUBLISH_BATCH_TARGET),
      Duration.fromString(submissionInterval.toString()),
      Duration.fromString(finalizedTime.toString()),
      new Date(),
      this.metadata
    )
    return new MsgCreateBridge(executor.key.accAddress, bridgeConfig)
  }

  MsgSetBridgeInfo(bridgeInfo: BridgeInfo) {
    return new MsgSetBridgeInfo(executorL2.key.accAddress, bridgeInfo)
  }

  async initialize() {
    const msgs = [
      this.MsgCreateBridge(this.submissionInterval, this.finalizedTime)
    ]

    const txRes = await sendTx(executor, msgs)

    // load bridge info from l1 chain and send to l2 chain
    let bridgeID = 0
    const txInfo = await config.l1lcd.tx.txInfo(txRes.txhash)
    for (const e of txInfo.events) {
      if (e.type !== 'create_bridge') {
        continue
      }

      for (const attr of e.attributes) {
        if (attr.key !== 'bridge_id') {
          continue
        }

        bridgeID = parseInt(attr.value, 10)
      }

      break
    }

    const bridgeInfo = await config.l1lcd.ophost.bridgeInfo(bridgeID)
    const l2Msgs = [this.MsgSetBridgeInfo(bridgeInfo)]

    await sendTx(executorL2, l2Msgs)
  }
}

async function main() {
  try {
    const initializer = new L2Initializer(
      config.SUBMISSION_INTERVAL,
      config.FINALIZATION_PERIOD,
      config.IBC_METADATA
    )
    console.log('=========Initializing L2=========')
    console.log('submissionInterval: ', initializer.submissionInterval)
    console.log('finalizedTime: ', initializer.finalizedTime)
    console.log('metadata: ', initializer.metadata)
    console.log('bridgeId: ', initializer.bridgeId)
    await initializer.initialize()
    console.log('=========L2 Initialized Done=========')
  } catch (e) {
    console.error(e)
  }
}

if (require.main === module) {
  main()
}
