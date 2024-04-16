import { Column, Entity, PrimaryColumn } from 'typeorm'

@Entity('record')
export default class RecordEntity {
  @PrimaryColumn()
  bridgeId: number

  @PrimaryColumn()
  batchIndex: number

  @Column()
  startBlockNumber: number

  @Column()
  endBlockNumber: number

  @Column('text', { array: true })
  batchInfo: string[] // for l1 => txHash, for celestia => height::commitment
}
