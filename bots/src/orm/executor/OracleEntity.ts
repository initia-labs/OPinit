import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('executor_oracle')
export default class ExecutorOracleEntity {
  @PrimaryColumn('int')
  blockHeight: number;

  @Column('int')
  blockTimestamp: Date;

  @Column('bigint')
  price: string;

  @Column('text')
  @Index('executor_oracle_pair_index')
  pair: string;
}