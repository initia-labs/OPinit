import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('coin')
export default class CoinEntity {
  @PrimaryColumn('text')
  l1StructTag: string;

  @Column('text')
  l1Denom: string;

  @Column('text')
  l2StructTag: string;

  @Column('text')
  @Index('coin_l2_denom')
  l2Denom: string;
}
