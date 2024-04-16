import { Column, Entity, PrimaryColumn } from 'typeorm'

@Entity('executor_output')
export default class OutputEntity {
  @PrimaryColumn('int')
  outputIndex: number

  @Column('text')
  outputRoot: string

  @Column('text')
  stateRoot: string

  @Column('text')
  merkleRoot: string

  @Column('text')
  lastBlockHash: string // last block hash of the epoch

  @Column('int')
  startBlockNumber: number // start block height of the epoch

  @Column('int')
  endBlockNumber: number // end block height of the epoch
}
