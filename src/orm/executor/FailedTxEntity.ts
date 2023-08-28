import { Column, Entity, PrimaryColumn } from 'typeorm';
import { Msg } from '@initia/initia.js';

@Entity('executor_failed_tx')
export default class FailedTxEntity {
  @PrimaryColumn('int')
  height: number;

  @PrimaryColumn('text')
  monitor: string;

  @Column({
    type: 'jsonb'
  })
  messages: string[];

  @Column({
    type: 'text',
    nullable: true
  })
  error: string;
}
