import {
  KoaController,
  Validate,
  Get,
  Controller,
  Validator,
} from 'koa-joi-controllers'
import { ErrorTypes } from 'lib/error'
import { success, error } from 'lib/response'
import { outputService } from 'service'

const Joi = Validator.Joi

@Controller('')
export class OutputController extends KoaController {
  @Get('/output/:output_index')
  @Validate({
    params: {
      output_index: Joi.number().description('Output index'),
    },
  })
  async getOutput(ctx): Promise<void> {
    const output = await outputService().getOutput(ctx.params.output_index)
    if (output) success(ctx, output)
    else error(ctx, ErrorTypes.NOT_FOUND_ERROR)
  }
}
