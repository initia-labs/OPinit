import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('tx')
export default class TxEntity {
  @PrimaryColumn('text')
  coinType: string;

  @PrimaryColumn('int')
  sequence: number;

  @Column('text')
  @Index('tx_sender_index')
  sender: string;

  @Column('text')
  @Index('tx_receiver_index')
  receiver: string;

  @Column('int')
  amount: number;

  @Column('int')
  @Index('tx_output_index')
  outputIndex: number;

  @Column('text')
  merkleRoot: string;

  @Column('text', { array: true })
  merkleProof: string[];
}
