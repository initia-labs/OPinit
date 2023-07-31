import { TxEntity } from 'orm'
import { Container, Service } from 'typedi'
import { Repository } from 'typeorm'
import { InjectRepository } from 'typeorm-typedi-extensions'

@Service()
export class TxService {
  constructor(
    @InjectRepository(TxEntity) private readonly repo: Repository<TxEntity>
  ) {}

  async getTx(
    coin_type: string,
    sequence: number
  ): Promise<TxEntity | null> {
    return this.repo.findOne({ where: { coin_type, sequence } })
  }
}

export function txService(): TxService {
  return Container.get(TxService)
}
