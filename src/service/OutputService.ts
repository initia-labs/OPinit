import { OutputEntity } from 'orm'
import { Container, Service } from 'typedi'
import { Repository } from 'typeorm'
import { InjectRepository } from 'typeorm-typedi-extensions'

@Service()
export class OutputService {
  constructor(
    @InjectRepository(OutputEntity)
    private readonly repo: Repository<OutputEntity>
  ) {}

  async getOutput(outputIndex: number): Promise<OutputEntity | null> {
    return this.repo.findOne({ where: { outputIndex } })
  }
}

export function outputService(): OutputService {
  return Container.get(OutputService)
}
