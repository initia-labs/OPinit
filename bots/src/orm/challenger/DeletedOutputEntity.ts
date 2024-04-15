import { Column, Entity, PrimaryColumn } from 'typeorm'

@Entity('challenger_deleted_output')
export default class ChallengedOutputEntity {
  @PrimaryColumn('bigint')
  outputIndex: number

  @Column('bigint')
  bridgeId: string

  @Column('text')
  reason: string
}
