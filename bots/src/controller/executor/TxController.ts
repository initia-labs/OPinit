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
import { getTx } from 'service';

const Joi = Validator.Joi;

@Controller('')
export class TxController extends KoaController {
  /**
   *
   * @api {get} /tx/:l1_metadata/:sequence Get tx entity
   * @apiName getTx
   * @apiGroup Tx
   *
   * @apiParam {String} l1Metadata L1 coin metadata
   * @apiParam {Number} sequence L2 withdrawal tx sequence
   *
   * @apiSuccess {String} l1Metadata L1 coin metadata
   * @apiSuccess {Number} sequence L2 sequence
   * @apiSuccess {String} sender   Withdrawal tx sender
   * @apiSuccess {String} receiver Withdrawal tx receiver
   * @apiSuccess {Number} amount   Withdrawal amount
   * @apiSuccess {Number} outputIndex Output index
   * @apiSuccess {String} merkleRoot Withdrawal tx merkle root
   * @apiSuccess {String[]} merkleProof Withdrawal txs merkle proof
   */
  @Get('/tx/:l1_metadata/:sequence')
  @Validate({
    params: {
      l1_metadata: Joi.string().description('L1 Metadata'),
      sequence: Joi.number().description('Sequence')
    },
    failure: ErrorCodes.INVALID_REQUEST_ERROR
  })
  async getTx(ctx: Context): Promise<void> {
    const l1_metadata: string = ctx.params.l1_metadata as string;
    const sequence: number = ctx.params.sequence as number;
    success(ctx, await getTx(l1_metadata, sequence));
  }
}
