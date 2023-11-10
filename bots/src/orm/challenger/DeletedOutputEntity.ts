import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('challenger_deleted_output')
export default class ChallengerDeletedOutputEntity {
  @PrimaryColumn('bigint')
  outputIndex: number;

  @Column('bigint')
  bridgeId: string;

  @Column('text')
  reason: string;
}
