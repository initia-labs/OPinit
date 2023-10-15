import { Wallet, MnemonicKey, BCS, Msg, MsgExecute } from '@initia/initia.js';
import { DataSource, ManyToMany } from 'typeorm';
import { getDB } from './db';
import {
  ChallengerWithdrawalTxEntity,
  ChallengerDepositTxEntity,
  ChallengerOutputEntity,
  StateEntity,
  ChallengerCoinEntity,
  DeletedOutputEntity
} from 'orm';
import { delay } from 'bluebird';
import { challengerLogger as logger } from 'lib/logger';
import { APIRequest } from 'lib/apiRequest';
import { GetLatestOutputResponse } from 'service';
import { fetchBridgeConfig } from 'lib/lcd';
import axios from 'axios';
import { GetAllCoinsResponse } from 'service/executor/CoinService';
import { getConfig } from 'config';
import { sendTx } from 'lib/tx';
import ChallengerHelper, { ENOT_EQUAL_TX } from './ChallegnerHelper';
import { EntityManager } from 'typeorm';

const config = getConfig();
const bcs = BCS.getInstance();

export class Challenger {
  private challenger: Wallet;
  private executor: Wallet;
  private isRunning = false;
  private db: DataSource;
  private apiRequester: APIRequest;
  private DEPOSIT_THRESHOLD = 10; // TODO: set threshold from contract config
  private WITHDRAWAL_THRESHOLD = 10; // TODO: set threshold from contract config
  helper: ChallengerHelper = new ChallengerHelper();

  constructor(public isFetch: boolean) {}

  async init() {
    // use to sync with bridge latest state
    if (this.isFetch) await this.fetchBridgeState();

    [this.db] = getDB();
    this.challenger = new Wallet(
      config.l1lcd,
      new MnemonicKey({ mnemonic: config.CHALLENGER_MNEMONIC })
    );
    this.executor = new Wallet(
      config.l1lcd,
      new MnemonicKey({ mnemonic: config.EXECUTOR_MNEMONIC })
    );
    this.isRunning = true;
  }

  // TODO: fetch from finalized state, not latest state
  public async fetchBridgeState() {
    [this.db] = getDB();
    this.apiRequester = new APIRequest(config.EXECUTOR_URI);
    const cfg = await fetchBridgeConfig();
    const outputRes = await this.apiRequester.getQuery<GetLatestOutputResponse>(
      '/output/latest'
    );
    if (!outputRes) return;
    const coinRes = await this.apiRequester.getQuery<GetAllCoinsResponse>(
      '/coin'
    );
    if (!coinRes) return;
    const l1Res = await axios.get(
      `${config.L1_LCD_URI}/cosmos/base/tendermint/v1beta1/blocks/latest`
    );
    if (!l1Res) return;

    await this.db.transaction(async (manager: EntityManager) => {
      await manager
        .getRepository(ChallengerOutputEntity)
        .save(outputRes.output);
      await manager.getRepository(ChallengerCoinEntity).save(coinRes.coins);
      await manager.getRepository(StateEntity).save([
        {
          name: 'challenger_l1_monitor',
          height: parseInt(l1Res.data.block.header.height)
        },
        {
          name: 'challenger_l2_monitor',
          height:
            outputRes.output.checkpointBlockHeight +
            Number.parseInt(cfg.submission_interval) -
            1
        }
      ]);
    });
  }

  public async run(): Promise<void> {
    await this.init();

    while (this.isRunning) {
      try {
        await this.l1Challenge();
        await this.l2Challenge();
      } catch (e) {
        this.stop();
      } finally {
        await delay(1_000);
      }
    }
  }

  public async l1Challenge() {
    await this.db.transaction(async (manager: EntityManager) => {
      const unchekcedDepositTx = await this.helper.getUncheckedTx(
        manager,
        ChallengerDepositTxEntity
      );
      if (!unchekcedDepositTx || !unchekcedDepositTx.finalizedOutputIndex)
        return;

      // case 1. not equal deposit tx between L1 and L2
      if (unchekcedDepositTx.finalizedOutputIndex === ENOT_EQUAL_TX) {
        await this.deleteL2Outptut(
          unchekcedDepositTx,
          'not same deposit tx between L1 and L2'
        );
        return;
      }

      // case2. not finalized within threshold
      if (
        unchekcedDepositTx.finalizedOutputIndex >
        unchekcedDepositTx.outputIndex + this.DEPOSIT_THRESHOLD
      ) {
        await this.deleteL2Outptut(
          unchekcedDepositTx,
          'deposit tx is not finalized for threshold submission interval'
        );
        return;
      }

      await this.helper.finalizeUncheckedTx(
        manager,
        ChallengerDepositTxEntity,
        unchekcedDepositTx
      );
    });
  }

  public stop(): void {
    this.isRunning = false;
  }

  async getChallengerOutputRoot(outputIndex: number): Promise<string | null> {
    const challengerOutputEntity = await this.db
      .getRepository(ChallengerOutputEntity)
      .find({
        where: { outputIndex: outputIndex },
        take: 1
      });

    if (challengerOutputEntity.length === 0) return null;
    return challengerOutputEntity[0].outputRoot;
  }

  async getContractOutputRoot(outputIndex: number): Promise<string | null> {
    try {
      const outputRootFromContract =
        await config.l1lcd.move.viewFunction<Uint8Array>(
          '0x1',
          'op_output',
          'get_output_root',
          [],
          [
            bcs.serialize(BCS.ADDRESS, this.executor.key.accAddress),
            bcs.serialize(BCS.STRING, config.L2ID),
            bcs.serialize(BCS.U64, outputIndex)
          ]
        );
      return Array.from(outputRootFromContract)
        .map((byte) => byte.toString(16))
        .join('');
    } catch (e) {
      logger.warn(
        `[L2 Challenger] waiting for submitting output root in output index ${outputIndex}`
      );
      return null;
    }
  }

  public async l2Challenge() {
    await this.db.transaction(async (manager: EntityManager) => {
      const uncheckedWithdrawalTx = await this.helper.getUncheckedTx(
        manager,
        ChallengerWithdrawalTxEntity
      );
      if (!uncheckedWithdrawalTx) return;

      // condition 1. ouptut should be submitted
      const lastIndex = await this.helper.getLastOutputIndex(
        manager,
        ChallengerOutputEntity
      );
      if (
        !uncheckedWithdrawalTx.outputIndex ||
        uncheckedWithdrawalTx.outputIndex > lastIndex
      )
        return;

      // case 1. output root not matched
      const outputRootFromContract = await this.getContractOutputRoot(
        uncheckedWithdrawalTx.outputIndex
      );
      const outputRootFromChallenger = await this.getChallengerOutputRoot(
        uncheckedWithdrawalTx.outputIndex
      );
      if (!outputRootFromContract || !outputRootFromChallenger) return;
      const isOutputFinalized = await this.isFinalizedOutput(
        uncheckedWithdrawalTx.outputIndex
      );
      if (
        !isOutputFinalized &&
        outputRootFromContract !== outputRootFromChallenger
      ) {
        await this.deleteL2Outptut(
          uncheckedWithdrawalTx,
          `not equal output root from contract: ${outputRootFromContract}, from challenger: ${outputRootFromChallenger}`
        );
        return;
      }

      await this.helper.finalizeUncheckedTx(
        manager,
        ChallengerWithdrawalTxEntity,
        uncheckedWithdrawalTx
      );
    });
  }

  async isFinalizedOutput(outputIndex: number) {
    const isFinalized: boolean = await config.l1lcd.move.viewFunction<boolean>(
      '0x1',
      'op_output',
      'is_finalized',
      [],
      [
        bcs.serialize(BCS.ADDRESS, this.executor.key.accAddress),
        bcs.serialize(BCS.STRING, config.L2ID),
        bcs.serialize(BCS.U64, outputIndex)
      ]
    );
    return isFinalized;
  }

  async isOutputSubmitted(outputIndex: number): Promise<boolean> {
    const nextBlockHeight = await config.l1lcd.move.viewFunction<string>(
      '0x1',
      'op_output',
      'next_block_num',
      [],
      [
        bcs.serialize('address', this.executor.key.accAddress),
        bcs.serialize('string', config.L2ID)
      ]
    );
    return parseInt(nextBlockHeight) > outputIndex;
  }

  async deleteL2Outptut(
    entity: ChallengerWithdrawalTxEntity | ChallengerDepositTxEntity,
    reason?: string
  ) {
    if (!(await this.isOutputSubmitted(entity.outputIndex))) return;

    const deletedOutput: DeletedOutputEntity = {
      outputIndex: entity.outputIndex,
      executor: this.executor.key.accAddress,
      l2Id: config.L2ID,
      reason: reason ?? 'unknown'
    };
    await this.db.getRepository(DeletedOutputEntity).save(deletedOutput);

    const executeMsg: Msg = new MsgExecute(
      this.challenger.key.accAddress,
      '0x1',
      'op_output',
      'delete_l2_output',
      [],
      [
        bcs.serialize('address', this.executor.key.accAddress),
        bcs.serialize('string', config.L2ID),
        bcs.serialize('u64', entity.outputIndex)
      ]
    );

    logger.info(
      `output index ${entity.outputIndex} is deleted, reason: ${reason}`
    );

    // await sendTx(this.challenger, [executeMsg]);
    // process.exit(0); // exit process when output is deleted
  }
}
