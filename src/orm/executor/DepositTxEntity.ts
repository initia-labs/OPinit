import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('executor_deposit_tx')
export default class DepositTxEntity {
  @PrimaryColumn('text')
  coinType: string;

  @PrimaryColumn('int')
  sequence: number;

  @Column('text')
  @Index('executor_deposit_tx_sender_index')
  sender: string;

  @Column('text')
  @Index('executor_deposit_tx_receiver_index')
  receiver: string;

  @Column('int')
  @Index('executor_deposit_tx_output_index')
  outputIndex: number;

  @Column('bigint')
  amount: number;
}
