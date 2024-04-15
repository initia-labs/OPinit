import { Column, Entity, PrimaryColumn } from 'typeorm'

@Entity('state')
export default class StateEntity {
  @PrimaryColumn('text')
  name: string

  @Column('int')
  height: number
}
