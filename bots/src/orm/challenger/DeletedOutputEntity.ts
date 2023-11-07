import { Column, Entity, Index, PrimaryColumn } from 'typeorm';

@Entity('challenger_deleted_output')
export default class DeletedOutputEntity {
  @PrimaryColumn('bigint')
  outputIndex: number;

  @Column('text')
  executor: string;

  @Column('text')
  l2Id: string;

  @Column('text')
  reason: string;
}
