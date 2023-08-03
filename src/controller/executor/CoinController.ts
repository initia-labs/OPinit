import {
  KoaController,
  Controller,
  Get,
  Validate,
  Validator
} from 'koa-joi-controllers';
import { success } from '../../lib/response';
import { getCoin } from '../../service/CoinService';
import { ErrorCodes } from 'lib/error';

const Joi = Validator.Joi;

@Controller('')
export default class CoinController extends KoaController {
  /**
   * @api {get} /coin/:coin_type Get coin mapping
   * @apiName getCoin
   * @apiGroup Coin
   *
   * @apiParam {String} coinType L1 coin type
   *
   * @apiSuccess {String} l1StructTag L1 coin struct tag
   * @apiSuccess {String} l1Denom L1 coin denom
   * @apiSuccess {String} l2StructTag L2 coin struct tag
   * @apiSuccess {String} l2Denom L2 coin denom
   *
   */
  @Get('/coin/:coin_type')
  @Validate({
    params: {
      coin_type: Joi.string().description('Coin type')
    },
    failure: ErrorCodes.INVALID_REQUEST_ERROR
  })
  async getCoin(ctx): Promise<void> {
    success(ctx, await getCoin(ctx.params.coin_type));
  }
}
