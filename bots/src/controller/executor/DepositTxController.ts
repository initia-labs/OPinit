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
import { getDepositTx } from 'service';

const Joi = Validator.Joi;

@Controller('')
export class DepositTxController extends KoaController {
  /**
   *
   * @api {get} /tx/:bridge_id/:sequence Get tx entity
   * @apiName getTx
   * @apiGroup Tx
   *
   * @apiParam {String} bridge_id L2 bridge id
   * @apiParam {Number} sequence L2 deposit tx sequence
   *
   * @apiSuccess {String} bridge_id L2 bridge id
   * @apiSuccess {Number} sequence L2 sequence
   * @apiSuccess {String} l1Denom Deposit coin L1 denom
   * @apiSuccess {String} l2Denom Deposit coin L2 denom
   * @apiSuccess {String} sender   Deposit tx sender
   * @apiSuccess {String} receiver Deposit tx receiver
   * @apiSuccess {Number} amount   Deposit amount
   * @apiSuccess {Number} outputIndex Output index
   * @apiSuccess {String} data Deposit tx data
   * @apiSuccess {Number} l1Height L1 height of deposit tx
   */
  @Get('/tx/deposit/:bridge_id/:sequence')
  @Validate({
    params: {
      bridge_id: Joi.string().description('L2 bridge id'),
      sequence: Joi.number().description('Sequence')
    },
    failure: ErrorCodes.INVALID_REQUEST_ERROR
  })
  async getDepositTx(ctx: Context): Promise<void> {
    const bridge_id: string = ctx.params.bridge_id as string;
    const sequence: number = ctx.params.sequence as number;
    success(ctx, await getDepositTx(bridge_id, sequence));
  }
}
