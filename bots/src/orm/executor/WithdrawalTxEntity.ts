import { Column, Entity, Index, PrimaryColumn } from 'typeorm'

@Entity('executor_withdrawal_tx')
export default class WithdrawalTxEntity {
  @PrimaryColumn('bigint')
  bridgeId: string

  @PrimaryColumn('bigint')
  sequence: string

  @Column('text')
  l1Denom: string

  @Column('text')
  l2Denom: string

  @Column('text')
  @Index('executor_withdrawal_tx_sender_index')
  sender: string

  @Column('text')
  @Index('executor_withdrawal_tx_receiver_index')
  receiver: string

  @Column('bigint')
  amount: string

  @Column('int')
  @Index('executor_withdrawal_tx_output_index')
  outputIndex: number

  @Column('text')
  merkleRoot: string

  @Column('text', { array: true })
  merkleProof: string[]
}
