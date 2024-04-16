import { Column, Entity, Index, PrimaryColumn } from 'typeorm'

@Entity('executor_unconfirmed_tx')
export default class UnconfirmedTxEntity {
  @PrimaryColumn('bigint')
  bridgeId: string

  @PrimaryColumn('bigint')
  sequence: string

  @Column('text')
  @Index('executor_unconfirmed_deposit_tx_sender_index')
  sender: string

  @Column('text')
  @Index('executor_unconfirmed_deposit_tx_receiver_index')
  receiver: string

  @Column('int')
  @Index('executor_unconfirmed_deposit_tx_output_index')
  outputIndex: number

  @Column('bigint')
  amount: string

  @Column('text')
  l1Denom: string

  @Column('text')
  l2Denom: string

  @Column('text')
  data: string

  @Column('int')
  l1Height: number

  @Column({
    type: 'text',
    nullable: true
  })
  error: string

  @Column()
  processed: boolean
}
