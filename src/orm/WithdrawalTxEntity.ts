import { Column, Entity, Index, PrimaryColumn } from 'typeorm'

@Entity('withdrawal_tx')
export default class WithdrawalTxEntity {
    @PrimaryColumn('int')
    height : number

    @Column('text')
    @Index('withdrawal_tx_from_index')
    from: string

    @Column('text')
    @Index('withdrawal_tx_to_index')
    to: string
    

    @Column('text')
    l2_token: string

    @Column('int')
    amount: number

    @Column('int')
    l2_sequence: number

    @Column('boolean')
    is_checked: boolean
}
