import { Context } from 'koa';
import {
  KoaController,
  Validate,
  Get,
  Controller,
  Validator
} from 'koa-joi-controllers';
import { ErrorCodes } from 'lib/error';
import { success } from 'lib/response';
import { getWithdrawalTx } from 'service';

const Joi = Validator.Joi;

@Controller('')
export class WithdrawalTxController extends KoaController {
  /**
   *
   * @api {get} /tx/:bridge_id/:sequence Get tx entity
   * @apiName getTx
   * @apiGroup Tx
   *
   * @apiParam {String} bridge_id L2 bridge id
   * @apiParam {Number} sequence L2 withdrawal tx sequence
   *
   * @apiSuccess {String} bridge_id L2 bridge id
   * @apiSuccess {Number} sequence L2 sequence
   * @apiSuccess {String} l1Denom Withdrawal coin L1 denom
   * @apiSuccess {String} l2Denom Withdrawal coin L2 denom
   * @apiSuccess {String} sender   Withdrawal tx sender
   * @apiSuccess {String} receiver Withdrawal tx receiver
   * @apiSuccess {Number} amount   Withdrawal amount
   * @apiSuccess {Number} outputIndex Output index
   * @apiSuccess {String} merkleRoot Withdrawal tx merkle root
   * @apiSuccess {String[]} merkleProof Withdrawal txs merkle proof
   */
  @Get('/tx/withdrawal/:bridge_id/:sequence')
  @Validate({
    params: {
      bridge_id: Joi.string().description('L2 bridge id'),
      sequence: Joi.number().description('Sequence')
    },
    failure: ErrorCodes.INVALID_REQUEST_ERROR
  })
  async getWithdrawalTx(ctx: Context): Promise<void> {
    const bridge_id: string = ctx.params.bridge_id as string;
    const sequence: number = ctx.params.sequence as number;
    success(ctx, await getWithdrawalTx(bridge_id, sequence));
  }
}
