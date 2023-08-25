import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('executor_coin')
export default class CoinEntity {
  @PrimaryColumn('text')
  l1StructTag: string;

  @Column('text')
  l1Denom: string;

  @Column('text')
  l2StructTag: string;

  @Column('text')
  @Index('executor_coin_l2_denom')
  l2Denom: string;
}
