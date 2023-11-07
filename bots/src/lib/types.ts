export interface BridgeConfig {
  submission_interval: string;
  challenger: string;
  proposer: string;
  finalization_period_seconds: string;
  starting_block_number: string;
}

export interface WithdrawalTx {
  sequence: number;
  sender: string;
  receiver: string;
  amount: number;
  l2_id: string;
  metadata: string;
}

export interface DepositTx {
  sequence: number;
  sender: string;
  receiver: string;
  amount: number;
  l2_id: string;
  l1_token: string;
  l2_token: string;
}

export interface L1TokenBridgeInitiatedEvent {
  from: string;
  to: string;
  l2_id: string;
  l1_token: string;
  l2_token: string;
  amount: number;
  l1_sequence: number;
}

export interface L2TokenBridgeInitiatedEvent {
  from: string;
  to: string;
  l2_token: string;
  amount: number;
  l2_sequence: number;
}
