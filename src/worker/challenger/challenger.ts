import config from 'config';
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
  WithdrawalTxEntity,
  DepositTxEntity,
  ChallengerOutputEntity
} from 'orm';
import { delay } from 'bluebird';
import { logger } from 'lib/logger';

const bcs = BCS.getInstance();

export class Challenger {
  private challenger: Wallet;
  private isRunning = false;
  private db: DataSource;

  async init() {
    [this.db] = getDB();
    this.challenger = new Wallet(
      config.l1lcd,
      new MnemonicKey({ mnemonic: config.CHALLENGER_MNEMONIC })
    );
  }

  public async run(): Promise<void> {
    while (!this.isRunning) {
      try {
        await this.l1Challenge();
        await this.l2Challenge();
      } catch (e) {
        console.log(e);
        logger.error('challenger error', e);
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

  // monitoring L1 deposit event and check the relayer works properly (L1 TokenBridgeInitiatedEvent)
  public async l1Challenge() {
    const unchekcedDepositTx = await this.getUncheckedDepositTx();
    if (!unchekcedDepositTx) return;

    const threshold = 1; // should be included in 1 submission interval

    if (!unchekcedDepositTx.finalizedOutputIndex) {
      const lastOutput = await this.getLastOutputFromDB();
      const lastIndex = lastOutput.length == 0 ? -1 : lastOutput[0].outputIndex;
      if (lastIndex > unchekcedDepositTx.outputIndex + threshold) {
        await this.deleteL2Ouptut(
          unchekcedDepositTx,
          'failed to check deposit tx'
        );
      }
      return;
    }

    if (
      unchekcedDepositTx.outputIndex + threshold <=
      unchekcedDepositTx.finalizedOutputIndex
    ) {
      logger.info(
        `successfully checked tx : coinType ${unchekcedDepositTx.coinType}, sequence ${unchekcedDepositTx.sequence}`
      );
      await this.checkDepositTx(unchekcedDepositTx);
      return;
    }

    await this.deleteL2Ouptut(unchekcedDepositTx, 'failed to check deposit tx');
  }

  async getUncheckedDepositTx(): Promise<DepositTxEntity | null> {
    const getUncheckedDepositTx: DepositTxEntity[] = await this.db
      .getRepository(DepositTxEntity)
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
  async getUncheckedWitdrawalTx(): Promise<WithdrawalTxEntity | null> {
    const uncheckedWithdrawalTx: WithdrawalTxEntity[] = await this.db
      .getRepository(WithdrawalTxEntity)
      .find({
        where: { isChecked: false },
        order: { sequence: 'ASC' },
        take: 1
      });

    if (uncheckedWithdrawalTx.length === 0) return null;
    return uncheckedWithdrawalTx[0];
  }

  async checkDepositTx(depositTxEntity: DepositTxEntity): Promise<void> {
    await this.db
      .getRepository(DepositTxEntity)
      .update(
        {
          coinType: depositTxEntity.coinType,
          sequence: depositTxEntity.sequence
        },
        { isChecked: true }
      );
  }

  async checkWithdrawalTx(
    withdrawalTxEntity: WithdrawalTxEntity
  ): Promise<void> {
    await this.db
      .getRepository(WithdrawalTxEntity)
      .update(
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
    const outputRootFromContract = await config.l1lcd.move.viewFunction<Buffer>(
      '0x1',
      'op_output',
      'get_output_root',
      [config.L2ID],
      [bcs.serialize(BCS.U64, outputIndex)]
    );
    return outputRootFromContract.toString('hex');
  }

  // monitoring L2 withdrawal event and check the relayer works properly (L2 TokenBridgeInitiatedEvent)
  public async l2Challenge() {
    const uncheckedWithdrawalTx = await this.getUncheckedWitdrawalTx();
    if (!uncheckedWithdrawalTx) return;

    const outputRootFromContract = await this.getContractOutputRoot(
      uncheckedWithdrawalTx.outputIndex
    );

    const outputRootFromChallenger = await this.getChallengerOutputRoot(
      uncheckedWithdrawalTx.outputIndex
    );

    if (!outputRootFromChallenger || !outputRootFromContract) return;

    if (outputRootFromContract === outputRootFromChallenger) {
      logger.info(
        `[L2 Challenger] output root matched for output index ${uncheckedWithdrawalTx.outputIndex}`
      );
      await this.checkWithdrawalTx(uncheckedWithdrawalTx);
      return;
    }

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

  async deleteL2Ouptut(
    entity: WithdrawalTxEntity | DepositTxEntity,
    reason?: string
  ) {
    const isFinalized = await this.isFinalizedL2Output(entity.outputIndex);

    if (isFinalized) {
      logger.warn(
        `[L2 Challenger] output index ${entity.outputIndex} is already finalized`
      );
      if (entity instanceof DepositTxEntity) await this.checkDepositTx(entity);
      if (entity instanceof WithdrawalTxEntity)
        await this.checkWithdrawalTx(entity);
      return;
    }

    const executeMsg: Msg = new MsgExecute(
      this.challenger.key.accAddress,
      '0x1',
      'op_output',
      'delete_l2_output',
      [config.L2ID],
      [bcs.serialize(BCS.U64, entity.outputIndex)]
    );

    await sendTx(config.l1lcd, this.challenger, [executeMsg]);
    logger.info(
      `[L2 Challenger] output index ${entity.outputIndex} is deleted, reason: ${reason}`
    );
    process.exit(1);
  }
}

/// Utils
async function sendTx(client: LCDClient, sender: Wallet, msg: Msg[]) {
  try {
    const signedTx = await sender.createAndSignTx({ msgs: msg });
    const broadcastResult = await client.tx.broadcast(signedTx);
    await checkTx(client, broadcastResult.txhash);
    return broadcastResult.txhash;
  } catch (error) {
    console.log(error);
    throw new Error(`Error in sendTx: ${error}`);
  }
}

export async function checkTx(
  lcd: LCDClient,
  txHash: string,
  timeout = 60000
): Promise<TxInfo | undefined> {
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeout) {
    try {
      const txInfo = await lcd.tx.txInfo(txHash);
      if (txInfo) return txInfo;
      await delay(1000);
    } catch (err) {
      throw new Error(`Failed to check transaction status: ${err.message}`);
    }
  }

  throw new Error('Transaction checking timed out');
}
