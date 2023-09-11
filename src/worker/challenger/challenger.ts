import {
  Wallet,
  MnemonicKey,
  BCS,
  LCDClient,
  Msg,
  TxInfo,
  MsgExecute
} from '@initia/initia.js';
import { DataSource } from 'typeorm';
import { getDB } from './db';
import {
  ChallengerWithdrawalTxEntity,
  ChallengerDepositTxEntity,
  ChallengerOutputEntity,
  StateEntity,
  ChallengerCoinEntity
} from 'orm';
import { delay } from 'bluebird';
import { challengerLogger as logger } from 'lib/logger';
import { APIRequest } from 'lib/apiRequest';
import { GetLatestOutputResponse } from 'service';
import { fetchBridgeConfig } from 'lib/lcd';
import axios from 'axios';
import { GetAllCoinsResponse } from 'service/CoinService';
import { getConfig } from 'config';
import { sendTx } from 'lib/tx';

const config = getConfig();
const bcs = BCS.getInstance();

export class Challenger {
  private challenger: Wallet;
  private isRunning = false;
  private db: DataSource;
  private apiRequester: APIRequest;
  private threshold = 0; // TODO: set threshold from contract config

  async init() {
    [this.db] = getDB();
    this.challenger = new Wallet(
      config.l1lcd,
      new MnemonicKey({ mnemonic: config.CHALLENGER_MNEMONIC })
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

    await this.db.getRepository(ChallengerOutputEntity).save(outputRes.output);
    await this.db.getRepository(ChallengerCoinEntity).save(coinRes.coins);
    await this.db.getRepository(StateEntity).save([
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
        await delay(1000);
      }
    }
  }

  public async getLastOutputFromDB(): Promise<ChallengerOutputEntity[]> {
    return await this.db.getRepository(ChallengerOutputEntity).find({
      order: { outputIndex: 'DESC' },
      take: 1
    });
  }

  public async getLastOutputIndex(): Promise<number> {
    const lastOutput = await this.getLastOutputFromDB();
    const lastIndex = lastOutput.length == 0 ? -1 : lastOutput[0].outputIndex;
    return lastIndex;
  }

  // monitoring L1 deposit event and check the relayer works properly
  public async l1Challenge() {
    // get unchecked deposit txs not finalized yet
    const unchekcedDepositTx = await this.getUncheckedDepositTx();
    if (!unchekcedDepositTx) return;
    if (unchekcedDepositTx.finalizedOutputIndex === -1) {
      await this.deleteL2Ouptut(
        unchekcedDepositTx,
        'not same between initialized and finalized deposit tx'
      );
      return;
    }

    if (!unchekcedDepositTx.finalizedOutputIndex) {
      const lastIndex = await this.getLastOutputIndex();
      // delete tx if it is not finalized for threshold submission interval
      if (lastIndex > unchekcedDepositTx.outputIndex + this.threshold) {
        await this.deleteL2Ouptut(
          unchekcedDepositTx,
          'deposit tx is not finalized for threshold submission interval'
        );
      }
      return;
    }

    if (
      unchekcedDepositTx.outputIndex + this.threshold <=
      unchekcedDepositTx.finalizedOutputIndex
    ) {
      logger.info(
        `[L1 Challenger] successfully checked tx : coinType ${unchekcedDepositTx.coinType}, sequence ${unchekcedDepositTx.sequence}`
      );
      await this.checkDepositTx(unchekcedDepositTx);
      return;
    }

    await this.deleteL2Ouptut(unchekcedDepositTx, 'failed to check deposit tx');
  }

  async getUncheckedDepositTx(): Promise<ChallengerDepositTxEntity | null> {
    const getUncheckedDepositTx: ChallengerDepositTxEntity[] = await this.db
      .getRepository(ChallengerDepositTxEntity)
      .find({
        where: { isChecked: false },
        order: { sequence: 'ASC' },
        take: 1
      });

    if (getUncheckedDepositTx.length === 0) return null;
    return getUncheckedDepositTx[0];
  }

  public stop(): void {
    this.isRunning = false;
  }

  // get unchecked withdrawal txs with smallest sequence from db
  async getUncheckedWitdrawalTx(): Promise<ChallengerWithdrawalTxEntity | null> {
    const uncheckedWithdrawalTx: ChallengerWithdrawalTxEntity[] = await this.db
      .getRepository(ChallengerWithdrawalTxEntity)
      .find({
        where: { isChecked: false },
        order: { sequence: 'ASC' },
        take: 1
      });

    if (uncheckedWithdrawalTx.length === 0) return null;
    return uncheckedWithdrawalTx[0];
  }

  async checkDepositTx(
    depositTxEntity: ChallengerDepositTxEntity
  ): Promise<void> {
    await this.db.getRepository(ChallengerDepositTxEntity).update(
      {
        coinType: depositTxEntity.coinType,
        sequence: depositTxEntity.sequence
      },
      { isChecked: true }
    );
  }

  async checkWithdrawalTx(
    withdrawalTxEntity: ChallengerWithdrawalTxEntity
  ): Promise<void> {
    await this.db.getRepository(ChallengerWithdrawalTxEntity).update(
      {
        coinType: withdrawalTxEntity.coinType,
        sequence: withdrawalTxEntity.sequence
      },
      { isChecked: true }
    );
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
          [config.L2ID],
          [bcs.serialize(BCS.U64, outputIndex)]
        );
      return Array.prototype.map
        .call(outputRootFromContract, (x) => x.toString(16).padStart(2, '0'))
        .join('');
    } catch {
      logger.info(
        '[L2 Challenger] waiting output submitter to submit output index : ',
        outputIndex
      );
      return null;
    }
  }

  // monitoring L2 withdrawal event and check the relayer works properly
  public async l2Challenge() {
    // get unchecked withdrawal txs from challenger
    const uncheckedWithdrawalTx = await this.getUncheckedWitdrawalTx();
    if (!uncheckedWithdrawalTx) return;
    const lastIndex = await this.getLastOutputIndex();
    if (uncheckedWithdrawalTx.outputIndex > lastIndex) return;

    // compare output root from contract and challenger
    const outputRootFromContract = await this.getContractOutputRoot(
      uncheckedWithdrawalTx.outputIndex
    );
    if (!outputRootFromContract) return;
    const outputRootFromChallenger = await this.getChallengerOutputRoot(
      uncheckedWithdrawalTx.outputIndex
    );

    if (!outputRootFromChallenger || !outputRootFromContract) return;
    if (outputRootFromContract === outputRootFromChallenger) {
      logger.info(
        `[L2 Challenger] successfully output root matched for output index ${uncheckedWithdrawalTx.outputIndex}`
      );

      await this.checkWithdrawalTx(uncheckedWithdrawalTx);
      return;
    }

    await this.checkIfFinalized(uncheckedWithdrawalTx);
    await this.deleteL2Ouptut(
      uncheckedWithdrawalTx,
      'failed to check withdrawal tx'
    );
  }

  async isFinalizedL2Output(outputIndex: number) {
    const isFinalized: boolean = await config.l1lcd.move.viewFunction<boolean>(
      '0x1',
      'op_output',
      'is_finalized',
      [config.L2ID],
      [bcs.serialize(BCS.U64, outputIndex)]
    );
    return isFinalized;
  }

  async checkIfFinalized(
    entity: ChallengerWithdrawalTxEntity | ChallengerDepositTxEntity
  ) {
    const isFinalized = await this.isFinalizedL2Output(entity.outputIndex);

    if (isFinalized) {
      logger.warn(
        `[L2 Challenger] output index ${entity.outputIndex} is already finalized`
      );
      if (entity instanceof ChallengerDepositTxEntity) {
        logger.warn(`[L2 Challenger] check deposit tx ${entity.sequence}`);
        await this.checkDepositTx(entity);
      } else if (entity instanceof ChallengerWithdrawalTxEntity) {
        logger.warn(`[L2 Challenger] check withdrawal tx ${entity.sequence}`);
        await this.checkWithdrawalTx(entity);
      }

      return;
    }
  }

  async deleteL2Ouptut(
    entity: ChallengerWithdrawalTxEntity | ChallengerDepositTxEntity,
    reason?: string
  ) {
    const executeMsg: Msg = new MsgExecute(
      this.challenger.key.accAddress,
      '0x1',
      'op_output',
      'delete_l2_output',
      [config.L2ID],
      [bcs.serialize(BCS.U64, entity.outputIndex)]
    );

    await sendTx(this.challenger, [executeMsg]);
    logger.info(
      `[L2 Challenger] output index ${entity.outputIndex} is deleted, reason: ${reason}`
    );
    process.exit(0);
  }
}
