import { BridgeInfo, MsgDeleteOutput } from '@initia/initia.js'
import { DataSource, MoreThan } from 'typeorm'
import { getDB } from './db'
import {
  ChallengerDepositTxEntity,
  ChallengerFinalizeDepositTxEntity,
  ChallengerOutputEntity,
  ChallengerWithdrawalTxEntity,
  ChallengedOutputEntity,
  ChallengeEntity
} from '../../orm'
import { delay } from 'bluebird'
import { challengerLogger as logger } from '../../lib/logger'
import { INTERVAL_MONITOR, config } from '../../config'
import { EntityManager } from 'typeorm'
import {
  getLastOutputInfo,
  getOutputInfoByIndex,
  getBridgeInfo
} from '../../lib/query'
import MonitorHelper from '../../lib/monitor/helper'
import winston from 'winston'
import { TxWallet, WalletType, getWallet, initWallet } from '../../lib/wallet'
import { buildChallengerNotification, notifySlack } from '../../lib/slack'

const THRESHOLD_MISS_INTERVAL = 5

export class Challenger {
  private isRunning = false
  private db: DataSource
  bridgeId: number
  bridgeInfo: BridgeInfo

  // members for challenge
  l1LastCheckedSequence: number
  l1DepositSequenceToCheck: number
  l2OutputIndexToCheck: number

  submissionIntervalMs: number
  missCount: number // count of miss interval to finalize deposit tx
  threshold: number // threshold of miss interval to finalize deposit tx
  helper: MonitorHelper
  challenger: TxWallet

  constructor(public logger: winston.Logger) {
    [this.db] = getDB()
    this.bridgeId = config.BRIDGE_ID
    this.isRunning = true
    this.missCount = 0

    this.helper = new MonitorHelper()
    initWallet(WalletType.Challenger, config.l1lcd)
    this.challenger = getWallet(WalletType.Challenger)
  }

  public name(): string {
    return 'challenge'
  }

  public stop(): void {
    this.isRunning = false
    process.exit()
  }

  async init(): Promise<void> {
    this.bridgeInfo = await getBridgeInfo(this.bridgeId)
    this.submissionIntervalMs =
      this.bridgeInfo.bridge_config.submission_interval.seconds.toNumber() *
      1000

    const state = await this.db.getRepository(ChallengeEntity).findOne({
      where: {
        name: this.name()
      }
    })

    if (!state) {
      await this.db.getRepository(ChallengeEntity).save({
        name: this.name(),
        l1DepositSequenceToCheck: 1,
        l1LastCheckedSequence: 0,
        l2OutputIndexToCheck: 1
      })
    }

    this.l1DepositSequenceToCheck = state?.l1DepositSequenceToCheck || 1
    this.l2OutputIndexToCheck = state?.l2OutputIndexToCheck || 1
    this.l1LastCheckedSequence = state?.l1LastCheckedSequence || 0
  }

  public async run(): Promise<void> {
    await this.init()
    while (this.isRunning) {
      try {
        await this.db.transaction(async (manager: EntityManager) => {
          await this.challengeDepositTx(manager)
          await this.challengeOutputRoot(manager)
        })
      } catch (err) {
        logger.error(`Challenger halted! ${err}`)
        this.stop()
      } finally {
        await delay(INTERVAL_MONITOR)
      }
    }
  }

  async challengeDepositTx(manager: EntityManager) {
    if (this.l1LastCheckedSequence == this.l1DepositSequenceToCheck) {
      // get next sequence from db with smallest sequence but bigger than last challenged sequence
      const nextDepositSequenceToCheck = await manager
        .getRepository(ChallengerDepositTxEntity)
        .find({
          where: { sequence: MoreThan(this.l1DepositSequenceToCheck) } as any,
          order: { sequence: 'ASC' },
          take: 1
        })

      if (nextDepositSequenceToCheck.length === 0) return
      this.l1DepositSequenceToCheck = Number(
        nextDepositSequenceToCheck[0].sequence
      )
    }

    const lastOutputInfo = await getLastOutputInfo(this.bridgeId)
    const depositTxFromChallenger = await manager
      .getRepository(ChallengerDepositTxEntity)
      .findOne({
        where: { sequence: this.l1DepositSequenceToCheck } as any
      })

    if (!depositTxFromChallenger) return
    this.l1DepositSequenceToCheck = Number(depositTxFromChallenger.sequence)

    // case 1. not finalized deposit tx
    const depositFinalizeTxFromChallenger = await manager
      .getRepository(ChallengerFinalizeDepositTxEntity)
      .findOne({
        where: { sequence: this.l1DepositSequenceToCheck } as any
      })

    if (!depositFinalizeTxFromChallenger) {
      this.missCount += 1
      this.logger.info(
        `[L1 Challenger] deposit tx with sequence "${this.l1DepositSequenceToCheck}" is not finialized`
      )
      if (this.missCount <= THRESHOLD_MISS_INTERVAL || !lastOutputInfo) {
        return await delay(this.submissionIntervalMs)
      }
      return await this.handleChallengedOutputProposal(
        manager,
        lastOutputInfo.output_index,
        `not finalized deposit tx within ${THRESHOLD_MISS_INTERVAL} submission interval ${depositFinalizeTxFromChallenger}`
      )
    }

    // case 2. not equal deposit tx between L1 and L2
    const pair = await config.l1lcd.ophost.tokenPairByL1Denom(
      this.bridgeId,
      depositTxFromChallenger.l1Denom
    )
    const isEqaul =
      depositTxFromChallenger.sender ===
        depositFinalizeTxFromChallenger.sender &&
      depositTxFromChallenger.receiver ===
        depositFinalizeTxFromChallenger.receiver &&
      depositTxFromChallenger.amount ===
        depositFinalizeTxFromChallenger.amount &&
      pair.l2_denom === depositFinalizeTxFromChallenger.l2Denom

    if (!isEqaul && lastOutputInfo) {
      await this.handleChallengedOutputProposal(
        manager,
        lastOutputInfo.output_index,
        `not equal deposit tx between L1 and L2`
      )
    }

    logger.info(
      `[L1 Challenger] deposit tx matched in sequence : ${this.l1DepositSequenceToCheck}`
    )

    this.missCount = 0
    this.l1LastCheckedSequence = this.l1DepositSequenceToCheck

    await manager.getRepository(ChallengeEntity).update(
      { name: this.name() },
      {
        l1DepositSequenceToCheck: this.l1DepositSequenceToCheck,
        l1LastCheckedSequence: this.l1LastCheckedSequence
      }
    )
  }

  async getChallengerOutputRoot(
    manager: EntityManager,
    outputIndex: number
  ): Promise<string | null> {
    const output = await getOutputInfoByIndex(this.bridgeId, outputIndex)
    if (!output) return null
    const startBlockNumber =
      outputIndex === 1
        ? 1
        : (await getOutputInfoByIndex(this.bridgeId, outputIndex - 1))
            .output_proposal.l2_block_number + 1
    const endBlockNumber = output.output_proposal.l2_block_number
    const blockInfo = await config.l2lcd.tendermint.blockInfo(endBlockNumber)

    const txEntities = await this.helper.getWithdrawalTxs(
      manager,
      ChallengerWithdrawalTxEntity,
      outputIndex
    )

    const merkleRoot = await this.helper.saveMerkleRootAndProof(
      manager,
      ChallengerWithdrawalTxEntity,
      txEntities
    )

    const outputEntity = this.helper.calculateOutputEntity(
      outputIndex,
      blockInfo,
      merkleRoot,
      startBlockNumber,
      endBlockNumber
    )

    await this.helper.saveEntity(manager, ChallengerOutputEntity, outputEntity)
    return outputEntity.outputRoot
  }

  async getContractOutputRoot(outputIndex: number): Promise<string | null> {
    try {
      const outputInfo = await config.l1lcd.ophost.outputInfo(
        this.bridgeId,
        outputIndex
      )
      return outputInfo.output_proposal.output_root
    } catch (err) {
      logger.info(
        `[L2 Challenger] waiting for submitting output root in output index ${outputIndex}`
      )
      return null
    }
  }

  async challengeOutputRoot(manager: EntityManager) {
    // condition 1. ouptut should be submitted
    const outputInfoToChallenge = await getOutputInfoByIndex(
      this.bridgeId,
      this.l2OutputIndexToCheck
    ).catch(() => {
      return null
    })

    if (!outputInfoToChallenge) return

    // case 1. output root not matched
    const outputRootFromContract = await this.getContractOutputRoot(
      this.l2OutputIndexToCheck
    )
    const outputRootFromChallenger = await this.getChallengerOutputRoot(
      manager,
      this.l2OutputIndexToCheck
    )

    if (!outputRootFromContract || !outputRootFromChallenger) return

    if (outputRootFromContract !== outputRootFromChallenger) {
      await this.handleChallengedOutputProposal(
        manager,
        this.l2OutputIndexToCheck,
        `not equal output root from contract: ${outputRootFromContract}, from challenger: ${outputRootFromChallenger}`
      )
    }

    logger.info(
      `[L2 Challenger] output root matched in output index : ${this.l2OutputIndexToCheck}`
    )
    this.l2OutputIndexToCheck += 1
    await manager.getRepository(ChallengeEntity).update(
      { name: this.name() },
      {
        l2OutputIndexToCheck: this.l2OutputIndexToCheck
      }
    )
  }

  async deleteOutputProposal(outputIndex: number) {
    const msg = new MsgDeleteOutput(
      this.challenger.key.accAddress,
      this.bridgeId,
      outputIndex
    )

    await this.challenger.transaction([msg])
  }

  async handleChallengedOutputProposal(
    manager: EntityManager,
    outputIndex: number,
    reason?: string
  ) {
    const challengedOutput: ChallengedOutputEntity = {
      outputIndex,
      bridgeId: this.bridgeId.toString(),
      reason: reason ?? 'unknown'
    }
    await manager.getRepository(ChallengedOutputEntity).save(challengedOutput)

    if (config.DELETE_OUTPUT_PROPOSAL === 'true') {
      await this.deleteOutputProposal(outputIndex)
    }

    await notifySlack(`${outputIndex}-${this.bridgeId}`, buildChallengerNotification(challengedOutput));
    process.exit();
  }
}
