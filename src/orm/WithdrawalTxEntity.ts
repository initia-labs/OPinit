import { Column, Entity, Index, PrimaryColumn } from 'typeorm'

@Entity('withdrawal_tx')
export default class WithdrawalTxEntity {
    @Column('int')
    height: number

    @PrimaryColumn('text')
    coin_type: string

    @PrimaryColumn('int')
    sequence: number
    
    @Column('text')
    @Index('withdrawal_tx_from_index')
    from: string

    @Column('text')
    @Index('withdrawal_tx_to_index')
    to: string
    
    @Column('int')
    amount: number

    @Column('boolean')
    is_checked: boolean
}
