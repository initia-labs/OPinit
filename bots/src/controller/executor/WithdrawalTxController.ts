import { Context } from 'koa'
import { KoaController, Get, Controller } from 'koa-joi-controllers'
import { ErrorTypes } from '../../lib/error'
import { error, success } from '../../lib/response'
import { getWithdrawalTxList } from '../../service'
import { responses, routeConfig, z } from 'koa-swagger-decorator'
import { GetWithdrawalResponse } from '../../swagger/executor_model'

@Controller('')
export class WithdrawalTxController extends KoaController {
  @routeConfig({
    method: 'get',
    path: '/tx/withdrawal',
    summary: 'Get withdrawal tx data',
    description: 'Get withdrawal data',
    tags: ['Withdrawal'],
    operationId: 'getWithdrawalTx',
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
  @responses(GetWithdrawalResponse)
  @Get('/tx/withdrawal')
  async getWithdrawalTxList(ctx: Context): Promise<void> {
    const withdrawalTxList = await getWithdrawalTxList(ctx.query as any)
    if (withdrawalTxList) success(ctx, withdrawalTxList)
    else error(ctx, ErrorTypes.API_ERROR)
  }
}
