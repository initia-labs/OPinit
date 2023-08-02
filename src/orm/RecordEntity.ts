import { Column, Entity, PrimaryColumn } from 'typeorm';

@Entity('record')
export default class RecordEntity {
  @PrimaryColumn()
  l2Id: string;

  @PrimaryColumn()
  batchIndex: number;

  @Column({
    type: 'bytea'
  })
  batch: Buffer;
}
