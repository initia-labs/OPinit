import { Column, Entity, Index, PrimaryColumn } from 'typeorm'

@Entity('challenger_deposit_tx')
export default class DepositTxEntity {
  @PrimaryColumn('bigint')
  sequence: string

  @Column('text')
  @Index('challenger_deposit_tx_sender_index')
  sender: string

  @Column('text')
  @Index('challenger_deposit_tx_receiver_index')
  receiver: string

  @Column('bigint')
  amount: string

  @Column('text')
  l1Denom: string

  @Column('text')
  l2Denom: string

  @Column('text')
  data: string
}
