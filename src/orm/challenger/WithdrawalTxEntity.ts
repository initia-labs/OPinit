import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('withdrawal_tx')
export default class WithdrawalTxEntity {
  @PrimaryColumn('text')
  coinType: string;

  @PrimaryColumn('int')
  sequence: number;

  @Column('text')
  @Index('withdrawal_tx_sender_index')
  sender: string;

  @Column('text')
  @Index('withdrawal_tx_receiver_index')
  receiver: string;

  @Column('int')
  amount: number;

  @Column('int')
  @Index('withdrawal_tx_output_index')
  outputIndex: number;

  @Column('text')
  merkleRoot: string;

  @Column('text', { array: true })
  merkleProof: string[];

  @Column('boolean')
  isChecked: boolean;
}