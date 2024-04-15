import { Column, Entity, Index, PrimaryColumn } from 'typeorm'

@Entity('challenger_finalize_withdrawal_tx')
export default class FinalizeWithdrawalTxEntity {
  @PrimaryColumn('bigint')
  bridgeId: string

  @PrimaryColumn('bigint')
  sequence: string

  @Column('text')
  l1Denom: string

  @Column('text')
  l2Denom: string

  @Column('text')
  @Index('challenger_finalize_tx_sender_index')
  sender: string

  @Column('text')
  @Index('challenger_finalize_tx_receiver_index')
  receiver: string

  @Column('bigint')
  amount: string

  @Column('int')
  @Index('challenger_finalize_tx_output_index')
  outputIndex: number
}
