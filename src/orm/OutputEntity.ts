import { Column, Entity, PrimaryColumn } from 'typeorm'

@Entity('output')
export default class OutputEntity {
  @PrimaryColumn('int')
  outputIndex: number

  @Column('text')
  outputRoot: string

  @Column('text')
  stateRoot: string

  @Column('text')
  storageRoot: string

  @Column('text')
  lastBlockHash: string

  @Column('int')
  startBlockHeight: number
}
