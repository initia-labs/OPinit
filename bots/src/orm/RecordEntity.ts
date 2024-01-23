import { Column, Entity, PrimaryColumn } from 'typeorm';

@Entity('record')
export default class RecordEntity {
  @PrimaryColumn()
  bridgeId: number;

  @PrimaryColumn()
  batchIndex: number;

  @Column()
  startBlockNumber: number;

  @Column()
  endBlockNumber: number;

  @Column({
    type: 'bytea'
  })
  batch: Buffer;
}
