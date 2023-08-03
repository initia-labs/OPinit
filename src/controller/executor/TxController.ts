import { Context } from 'koa';
import {
  KoaController,
  Validate,
  Get,
  Controller,
  Validator
} from 'koa-joi-controllers';
import { success } from 'lib/response';
import { getTx } from 'service';

const Joi = Validator.Joi;

@Controller('')
export class TxController extends KoaController {
  /**
   *
   * @api {get} /tx/:coin_type/:sequence Get tx entity
   * @apiName getTx
   * @apiGroup Tx
   *
   * @apiParam {String} coinType L1 coin struct tag
   * @apiParam {Number} sequence L2 withdrawal tx sequence
   *
   * @apiSuccess {String} coinType L1 coin struct tag
   * @apiSuccess {Number} sequence L2 sequence
   * @apiSuccess {String} sender   Withdrawal tx sender
   * @apiSuccess {String} receiver Withdrawal tx receiver
   * @apiSuccess {Number} amount   Withdrawal amount
   * @apiSuccess {Number} outputIndex Output index
   * @apiSuccess {String} merkleRoot Withdrawal tx merkle root
   * @apiSuccess {String[]} merkleProof Withdrawal txs merkle proof
   */
  @Get('/tx/:coin_type/:sequence')
  @Validate({
    params: {
      coin_type: Joi.string().description('Coin type'),
      sequence: Joi.number().description('Sequence')
    }
  })
  async getTx(ctx: Context): Promise<void> {
    const coin_type: string = ctx.params.coin_type as string;
    const sequence: number = ctx.params.sequence as number;
    success(ctx, await getTx(coin_type, sequence));
  }
}
