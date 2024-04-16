import { Column, Entity, PrimaryColumn } from 'typeorm'

@Entity('challenge')
export default class ChallengeEntity {
  @PrimaryColumn('text')
  name: string

  @Column('int')
  l1DepositSequenceToCheck: number

  @Column('int')
  l1LastCheckedSequence: number

  @Column('int')
  l2OutputIndexToCheck: number
}
