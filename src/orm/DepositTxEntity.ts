import { Column, Entity, Index, PrimaryColumn } from 'typeorm'

@Entity('deposit_tx')
export default class DepositTxEntity {

    @PrimaryColumn('int')
    height : number
    
    @Column('int')
    l1_sequence: number

    @Column('text')
    @Index('deposit_tx_from_index')
    from: string

    @Column('text')
    @Index('deposit_tx_to_index')
    to: string

    @Column('text')
    l2_id: string

    @Column('text')
    l1_token: string

    @Column('text')
    l2_token: string

    @Column('int')
    amount: number

    @Column('boolean')
    is_checked: boolean
}
