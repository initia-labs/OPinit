import {
  KoaController,
  Controller,
  Get,
  Validate,
  Validator
} from 'koa-joi-controllers';
import { success } from '../../lib/response';
import { getCoin, getAllCoins } from '../../service/executor/CoinService';
import { ErrorCodes } from 'lib/error';

const Joi = Validator.Joi;

@Controller('')
export default class CoinController extends KoaController {
  /**
   * @api {get} /coin Get all coin mapping
   * @apiName getAllCoins
   * @apiGroup Coin
   *
   * @apiSuccess {Object[]} coins Coin mapping list
   */
  @Get('/coin')
  async getAllCoins(ctx): Promise<void> {
    success(ctx, await getAllCoins());
  }

  /**
   * @api {get} /coin/:metadata Get coin mapping
   * @apiName getCoin
   * @apiGroup Coin
   *
   * @apiParam {String} l1Metadata L1 coin metadata
   *
   * @apiSuccess {String} l1Metadata L1 coin metadata
   * @apiSuccess {String} l1Denom L1 coin denom
   * @apiSuccess {String} l2Metadata L2 coin metadata
   * @apiSuccess {String} l2Denom L2 coin denom
   *
   */
  @Get('/coin/:metadata')
  @Validate({
    params: {
      metadata: Joi.string().description('Coin type')
    },
    failure: ErrorCodes.INVALID_REQUEST_ERROR
  })
  async getCoin(ctx): Promise<void> {
    success(ctx, await getCoin(ctx.params.metadata));
  }
}
