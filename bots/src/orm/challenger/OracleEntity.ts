import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('challenger_oracle')
export default class ChallengerOracleEntity {
  @PrimaryColumn('int')
  blockHeight: number;

  @Column('int')
  blockTimestamp: Date;

  @Column('bigint')
  price: string;

  @Column('text')
  @Index('challenger_oracle_pair_index')
  pair: string;
}