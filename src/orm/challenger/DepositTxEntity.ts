import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('deposit_tx')
export default class DepositTxEntity {
  @PrimaryColumn('text')
  coinType: string;

  @PrimaryColumn('int')
  sequence: number;

  @Column('text')
  @Index('deposit_tx_sender_index')
  sender: string;

  @Column('text')
  @Index('deposit_tx_receiver_index')
  receiver: string;

  @Column('int')
  @Index('deposit_tx_output_index')
  outputIndex: number;

  @Column('int')
  amount: number;

  @Column('boolean')
  isChecked: boolean;
}
