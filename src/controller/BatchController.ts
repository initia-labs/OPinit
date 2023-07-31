import { KoaController, Controller, Get } from "koa-joi-controllers";
import { success } from '../lib/response'
import { getBatch } from '../service/BatchService'

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
  @Get('/batch/:index')
  async getBatch(ctx): Promise<void> {
    success(ctx, await getBatch(ctx.params.index))
  }
}