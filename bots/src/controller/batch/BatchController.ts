import {
  KoaController,
  Controller,
  Get,
  Validate,
  Validator
} from 'koa-joi-controllers'
import { responses, routeConfig, z } from 'koa-swagger-decorator'
import { success } from '../../lib/response'
import { GetBatchResponse } from 'sawgger/batch_model'
import { getBatch } from 'service/batch/BatchService'

const Joi = Validator.Joi

@Controller('')
export class BatchController extends KoaController {
  @routeConfig({
    method: 'get',
    path: '/batch/{batch_index}',
    summary: 'Get batch',
    description: 'Get batch',
    tags: ['Batch'],
    operationId: 'getBatch',
    request: {
      params: z.object({
        batch_index: z.number()
      })
    }
  })
  @responses(GetBatchResponse)
  @Get('/batch/{batch_index}')
  @Validate({
    params: {
      batch_index: Joi.number().description('Batch index')
    }
  })
  async getBatch(ctx): Promise<void> {
    success(ctx, await getBatch(ctx.params.batch_index))
  }
}
