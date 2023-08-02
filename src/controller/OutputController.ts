import {
  KoaController,
  Validate,
  Get,
  Controller,
  Validator
} from 'koa-joi-controllers';
import { success } from 'lib/response';
import { getOutput } from 'service';

const Joi = Validator.Joi;

@Controller('')
export class OutputController extends KoaController {
  /**
   *
   * @api {get} /output Get output entity
   * @apiName getOutput
   * @apiGroup Output
   *
   * @apiParam {Number} [outputIndex] Index for output
   */
  @Get('/output/:output_index')
  @Validate({
    params: {
      output_index: Joi.number().description('Output index')
    }
  })
  async getOutput(ctx): Promise<void> {
    success(ctx, await getOutput(ctx.params.output_index));
  }
}
