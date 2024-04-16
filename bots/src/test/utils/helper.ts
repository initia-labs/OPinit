import {
  Wallet,
  MnemonicKey,
  BCS,
  Msg,
  MsgFinalizeTokenWithdrawal,
  Coin
} from '@initia/initia.js'

import { config } from '../../config'
import { sha3_256 } from '../../lib/util'
import { ExecutorOutputEntity } from '../../orm/index'
import WithdrawalTxEntity from '../../orm/executor/WithdrawalTxEntity'

export const bcs = BCS.getInstance()
export const executor = new Wallet(
  config.l1lcd,
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

export async function makeFinalizeMsg(
  txRes: WithdrawalTxEntity,
  outputRes: ExecutorOutputEntity
): Promise<Msg> {
  const msg = new MsgFinalizeTokenWithdrawal(
    config.BRIDGE_ID,
    outputRes.outputIndex,
    txRes.merkleProof,
    txRes.sender,
    txRes.receiver,
    parseInt(txRes.sequence),
    new Coin('uinit', txRes.amount),
    sha3_256(outputRes.outputIndex).toString('base64'),
    outputRes.stateRoot,
    outputRes.outputRoot,
    outputRes.lastBlockHash
  )
  return msg
}
