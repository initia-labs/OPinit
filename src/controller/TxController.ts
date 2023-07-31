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
import { txService } from 'service'

const Joi = Validator.Joi

@Controller('')
export class TxController extends KoaController {
  @Get('/tx/:sequence')
  @Validate({
    query: {
      coin_type: Joi.string().description('Coin type'),
      sequence: Joi.number().description('Sequence'),
    },
  })
  async getTx(ctx: Context): Promise<void> {
    const coin_type: string = ctx.query.coin_type as string
    const sequence: number = ctx.query.sequence as number
    const tx = await txService().getTx(coin_type, sequence)
    if (tx) success(ctx, tx)
    else error(ctx, ErrorTypes.NOT_FOUND_ERROR)
  }
}
