import { Column, Entity, PrimaryColumn } from 'typeorm';

@Entity('challenge')
export default class ChallengeEntity {
  @PrimaryColumn('text')
  name: string;

  @Column('int')
  l1DepositSequenceToChallenge: number;

  @Column('int')
  l2OutputIndexToChallenge: number;
}
