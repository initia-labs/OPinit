import { Column, Entity, Index, PrimaryColumn } from 'typeorm'

@Entity('challenger_withdrawal_tx')
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
  @Index('challenger_tx_sender_index')
  sender: string

  @Column('text')
  @Index('challenger_tx_receiver_index')
  receiver: string

  @Column('bigint')
  amount: string

  @Column('int')
  @Index('challenger_tx_output_index')
  outputIndex: number

  @Column('text')
  merkleRoot: string

  @Column('text', { array: true })
  merkleProof: string[]
}
