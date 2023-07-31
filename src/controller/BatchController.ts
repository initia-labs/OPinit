import { KoaController, Controller, Get, Validate, Validator } from "koa-joi-controllers";
import { success } from '../lib/response'
import { getBatch } from '../service/BatchService'

const Joi = Validator.Joi

@Controller('')
export default class BatchController extends KoaController {
    /**
   * @api {get} /batch Get batch
   * @apiName getBatch
   * @apiGroup Batch
   *
   * @apiParam {Number} [batchIndex] Index for batch
   *
   */
  @Get('/batch/:batch_index')
  @Validate({
    params: {
      batch_index: Joi.number().description('Batch index'),
    },
  })
  async getBatch(ctx): Promise<void> {
    success(ctx, await getBatch(ctx.params.batch_index))
  }
}