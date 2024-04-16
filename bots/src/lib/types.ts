import DepositTxEntity from '../orm/executor/DepositTxEntity'
import WithdrawalTxEntity from '../orm/executor/WithdrawalTxEntity'
import { ExecutorOutputEntity } from '../orm/index'

export interface WithdrawalTx {
  bridge_id: bigint;
  sequence: bigint;
  sender: string;
  receiver: string;
  l1_denom: string;
  amount: bigint;
}

/// response types

export interface WithdrawalTxResponse {
  withdrawalTx: WithdrawalTxEntity;
}

export interface DepositTxResponse {
  depositTx: DepositTxEntity;
}

export interface OutputResponse {
  output: ExecutorOutputEntity;
}
