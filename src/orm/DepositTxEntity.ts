import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('deposit_tx')
export default class DepositTxEntity {
  @Column('int')
  height: number;

  @PrimaryColumn('text')
  coin_type: string;

  @PrimaryColumn('int')
  sequence: number;

  @Column('text')
  @Index('deposit_tx_from_index')
  from: string;

  @Column('text')
  @Index('deposit_tx_to_index')
  to: string;

  @Column('text')
  l2_id: string;

  @Column('text')
  l2_token: string;

  @Column('int')
  amount: number;

  @Column('boolean')
  is_checked: boolean;
}
