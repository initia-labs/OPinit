import { ExecutorWithdrawalTxEntity, ExecutorOutputEntity } from '../../orm'
import { getDB } from '../../lib/db'
import { APIError, ErrorTypes } from '../../lib/error'
import { sha3_256 } from '../../lib/util'

interface ClaimTx {
  bridgeId: number;
  outputIndex: number;
  merkleProof: string[];
  sender: string;
  receiver: string;
  amount: number;
  l2Denom: string;
  version: string;
  stateRoot: string;
  merkleRoot: string;
  lastBlockHash: string;
}

export interface GetClaimTxListParam {
  sequence?: number;
  address?: string;
  stage?: number;
  offset?: number;
  limit: number;
  descending: string;
}

export interface GetClaimTxListResponse {
  count?: number;
  next?: number;
  limit: number;
  claimTxList: ClaimTx[];
}

export async function getClaimTxList(
  param: GetClaimTxListParam
): Promise<GetClaimTxListResponse> {
  const [db] = getDB()
  const queryRunner = db.createQueryRunner('slave')

  try {
    const offset = param.offset ?? 0
    const order = param.descending == 'true' ? 'DESC' : 'ASC'
    const claimTxList: ClaimTx[] = []

    const withdrawalQb = queryRunner.manager.createQueryBuilder(
      ExecutorWithdrawalTxEntity,
      'tx'
    )

    if (param.address) {
      withdrawalQb.andWhere('tx.sender = :sender', { sender: param.address })
    }

    if (param.sequence) {
      withdrawalQb.andWhere('tx.sequence = :sequence', {
        sequence: param.sequence
      })
    }

    const withdrawalTxs = await withdrawalQb
      .orderBy('tx.sequence', order)
      .skip(offset * param.limit)
      .take(param.limit)
      .getMany()

    withdrawalTxs.map(async (withdrawalTx) => {
      const outputQb = queryRunner.manager
        .createQueryBuilder(ExecutorOutputEntity, 'output')
        .where('output.output_index = :outputIndex', {
          outputIndex: withdrawalTx.outputIndex
        })

      const output = await outputQb.getOne()

      if (!output) {
        throw new APIError(ErrorTypes.NOT_FOUND_ERROR)
      }

      const claimData: ClaimTx = {
        bridgeId: parseInt(withdrawalTx.bridgeId),
        outputIndex: withdrawalTx.outputIndex,
        merkleProof: withdrawalTx.merkleProof,
        sender: withdrawalTx.sender,
        receiver: withdrawalTx.receiver,
        amount: parseInt(withdrawalTx.amount),
        l2Denom: withdrawalTx.l2Denom,
        version: sha3_256(withdrawalTx.outputIndex).toString('base64'),
        stateRoot: output.stateRoot,
        merkleRoot: output.merkleRoot,
        lastBlockHash: output.lastBlockHash
      }
      claimTxList.push(claimData)
    })

    const count = await withdrawalQb.getCount()
    let next: number | undefined

    if (count > (offset + 1) * param.limit) {
      next = offset + 1
    }

    return {
      count,
      next,
      limit: param.limit,
      claimTxList
    }
  } finally {
    await queryRunner.release()
  }
}
