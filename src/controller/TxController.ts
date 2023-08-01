import { Context } from 'koa'
import {
  KoaController,
  Validate,
  Get,
  Controller,
  Validator,
} from 'koa-joi-controllers'
import { ErrorTypes } from 'lib/error'
import { success, error } from 'lib/response'
import { getTx } from 'service'

const Joi = Validator.Joi

@Controller('')
export class TxController extends KoaController {

  /**
   * 
   * @api {get} /tx Get tx entity
   * @apiName getTx
   * @apiGroup Tx
   * 
   * @apiParam {String} [coin_type] Coin type
   * @apiParam {Number} [sequence] Sequence
   * 
   */
  @Get('/tx/:coin_type/:sequence')
  @Validate({
    query: {
      coin_type: Joi.string().description('Coin type'),
      sequence: Joi.number().description('Sequence'),
    },
  })
  async getTx(ctx: Context): Promise<void> {
    const coin_type: string = ctx.query.coin_type as string
    const sequence: number = ctx.query.sequence as number
    success(ctx, await getTx(coin_type, sequence))
  }
}
