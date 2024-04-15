import { Column, Entity, Index, PrimaryColumn } from 'typeorm'

@Entity('challenger_finalize_deposit_tx')
export default class FinalizeDepositTxEntity {
  // l1 sequence
  @PrimaryColumn('bigint')
  sequence: string

  @Column('text')
  @Index('challenger_finalize_deposit_tx_sender_index')
  sender: string

  @Column('text')
  @Index('challenger_finalize_deposit_tx_receiver_index')
  receiver: string

  @Column('bigint')
  amount: string

  @Column('text')
  l2Denom: string

  @Column('int')
  l1Height: number
}
