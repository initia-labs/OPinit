import { ExecutorOutputEntity, ExecutorWithdrawalTxEntity } from '../../orm'
import { Monitor } from './monitor'
import { WithdrawStorage } from '../storage'
import { WithdrawalTx } from '../types'
import { EntityManager } from 'typeorm'
import { BlockInfo } from '@initia/initia.js'
import { getDB } from '../../worker/bridgeExecutor/db'
import { RPCClient, RPCSocket } from '../rpc'
import winston from 'winston'
import { config } from '../../config'
import { getBridgeInfo } from '../query'

export class L2Monitor extends Monitor {
  submissionInterval: number
  nextSubmissionTimeSec: number

  constructor(
    public socket: RPCSocket,
    public rpcClient: RPCClient,
    logger: winston.Logger
  ) {
    super(socket, rpcClient, logger);
    [this.db] = getDB()
    this.nextSubmissionTimeSec = this.getCurTimeSec()
  }

  public name(): string {
    return 'executor_l2_monitor'
  }

  dateToSeconds(date: Date): number {
    return Math.floor(date.getTime() / 1000)
  }

  private async setNextSubmissionTimeSec(): Promise<void> {
    const bridgeInfo = await getBridgeInfo(this.bridgeId)
    this.submissionInterval =
      bridgeInfo.bridge_config.submission_interval.seconds.toNumber()
    this.nextSubmissionTimeSec += this.submissionInterval
  }

  private getCurTimeSec(): number {
    return this.dateToSeconds(new Date())
  }

  private async handleInitiateTokenWithdrawalEvent(
    manager: EntityManager,
    data: { [key: string]: string }
  ): Promise<void> {
    const outputInfo = await this.helper.getLastOutputFromDB(
      manager,
      ExecutorOutputEntity
    )
    if (!outputInfo) return
    const pair = await config.l1lcd.ophost.tokenPairByL2Denom(
      this.bridgeId,
      data['denom']
    )

    const tx: ExecutorWithdrawalTxEntity = {
      l1Denom: pair.l1_denom,
      l2Denom: pair.l2_denom,
      sequence: data['l2_sequence'],
      sender: data['from'],
      receiver: data['to'],
      amount: data['amount'],
      bridgeId: this.bridgeId.toString(),
      outputIndex: outputInfo ? outputInfo.outputIndex + 1 : 1,
      merkleRoot: '',
      merkleProof: []
    }

    await this.helper.saveEntity(manager, ExecutorWithdrawalTxEntity, tx)
  }

  public async handleEvents(manager: EntityManager): Promise<boolean> {
    const [isEmpty, events] = await this.helper.fetchAllEvents(
      config.l2lcd,
      this.currentHeight
    )
    if (isEmpty) return false

    const withdrawalEvents = events.filter(
      (evt) => evt.type === 'initiate_token_withdrawal'
    )
    for (const evt of withdrawalEvents) {
      const attrMap = this.helper.eventsToAttrMap(evt)
      await this.handleInitiateTokenWithdrawalEvent(manager, attrMap)
    }

    return true
  }

  private async saveMerkleRootAndProof(
    manager: EntityManager,
    entities: ExecutorWithdrawalTxEntity[]
  ): Promise<string> {
    const txs: WithdrawalTx[] = entities.map(
      (entity: ExecutorWithdrawalTxEntity) => ({
        bridge_id: BigInt(entity.bridgeId),
        sequence: BigInt(entity.sequence),
        sender: entity.sender,
        receiver: entity.receiver,
        l1_denom: entity.l1Denom,
        amount: BigInt(entity.amount)
      })
    )

    const storage = new WithdrawStorage(txs)
    const merkleRoot = storage.getMerkleRoot()
    for (let i = 0; i < entities.length; i++) {
      entities[i].merkleRoot = merkleRoot
      entities[i].merkleProof = storage.getMerkleProof(txs[i])
      await this.helper.saveEntity(
        manager,
        ExecutorWithdrawalTxEntity,
        entities[i]
      )
    }
    return merkleRoot
  }

  public async handleBlock(manager: EntityManager): Promise<void> {
    if (this.getCurTimeSec() < this.nextSubmissionTimeSec) return
    const lastOutput = await this.helper.getLastOutputFromDB(
      manager,
      ExecutorOutputEntity
    )

    const lastOutputEndBlockNumber = lastOutput ? lastOutput.endBlockNumber : 0
    const lastOutputIndex = lastOutput ? lastOutput.outputIndex : 0

    const startBlockNumber = lastOutputEndBlockNumber + 1
    const endBlockNumber = this.currentHeight
    const outputIndex = lastOutputIndex + 1

    if (startBlockNumber > endBlockNumber) return

    const blockInfo: BlockInfo = await config.l2lcd.tendermint.blockInfo(
      this.currentHeight
    )

    // fetch txs and build merkle tree for withdrawal storage
    const txEntities = await this.helper.getWithdrawalTxs(
      manager,
      ExecutorWithdrawalTxEntity,
      outputIndex
    )

    const merkleRoot = await this.saveMerkleRootAndProof(manager, txEntities)

    const outputEntity = this.helper.calculateOutputEntity(
      outputIndex,
      blockInfo,
      merkleRoot,
      startBlockNumber,
      endBlockNumber
    )

    await this.helper.saveEntity(manager, ExecutorOutputEntity, outputEntity)

    await this.setNextSubmissionTimeSec()
  }
}
