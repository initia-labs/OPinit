import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('challenger_coin')
export default class ChallengerCoinEntity {
  @PrimaryColumn('text')
  l1StructTag: string;

  @Column('text')
  l1Denom: string;

  @Column('text')
  l2StructTag: string;

  @Column('text')
  @Index('challenger_coin_l2_denom')
  l2Denom: string;
}
