import { Context } from 'koa'
import { KoaController, Get, Controller } from 'koa-joi-controllers'
import { ErrorTypes } from '../../lib/error'
import { error, success } from '../../lib/response'
import { getDepositTxList } from '../../service'
import { responses, routeConfig, z } from 'koa-swagger-decorator'
import { GetDepositResponse } from '../../swagger/executor_model'

@Controller('')
export class DepositTxController extends KoaController {
  @routeConfig({
    method: 'get',
    path: '/tx/deposit',
    summary: 'Get deposit tx data',
    description: 'Get deposit data',
    tags: ['Deposit'],
    operationId: 'getDepositTx',
    request: {
      query: z.object({
        address: z.string().optional(),
        sequence: z.number().optional(),
        limit: z
          .number()
          .optional()
          .default(20)
          .refine((value) => [10, 20, 100, 500].includes(value), {
            message: 'Invalid limit value'
          }),
        offset: z.number().optional().default(0),
        descending: z.boolean().optional().default(true)
      })
    }
  })
  @responses(GetDepositResponse)
  @Get('/tx/deposit')
  async getDepositTxList(ctx: Context): Promise<void> {
    const depositTxList = await getDepositTxList(ctx.query as any)
    if (depositTxList) success(ctx, depositTxList)
    else error(ctx, ErrorTypes.API_ERROR)
  }
}
