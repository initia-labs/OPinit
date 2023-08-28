import { Column, Entity, PrimaryColumn } from 'typeorm';

@Entity('executor_output')
export default class OutputEntity {
  @PrimaryColumn('int')
  outputIndex: number;

  @Column('text')
  outputRoot: string;

  @Column('text')
  stateRoot: string;

  @Column('text')
  storageRoot: string;

  @Column('text')
  lastBlockHash: string; // last block hash of the epoch

  @Column('int')
  checkpointBlockHeight: number; // start block height of the epoch
}
