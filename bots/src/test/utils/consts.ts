import { MnemonicKey } from '@initia/initia.js'
import { config } from '../../config'
import { TxWallet } from '../../lib/wallet'

export const { DEPOSITOR_MNEMONIC } = process.env

export const L1_SENDER = new TxWallet(
  config.l1lcd,
  new MnemonicKey({
    mnemonic: DEPOSITOR_MNEMONIC
  })
)

export const L2_RECEIVER = new TxWallet(
  config.l2lcd,
  new MnemonicKey({
    mnemonic: DEPOSITOR_MNEMONIC
  })
)
