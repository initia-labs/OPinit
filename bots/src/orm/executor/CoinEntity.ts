import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('executor_coin')
export default class CoinEntity {
  @PrimaryColumn('text')
  l1Metadata: string;

  @Column('text')
  l1Denom: string;

  @Column('text')
  l2Metadata: string;

  @Column('text')
  @Index('executor_coin_l2_denom')
  l2Denom: string;

  @Column('boolean')
  isChecked: boolean;
}
