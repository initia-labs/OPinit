import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('challenger_deposit_tx')
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

  @Column('int', { nullable: true })
  @Index('deposit_tx_finalized_output_index')
  finalizedOutputIndex: number | null;

  @Column('bigint')
  amount: number;

  @Column('boolean')
  isChecked: boolean;
}
