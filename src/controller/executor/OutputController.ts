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
   * @api {get} /output/:output_index Get output entity
   * @apiName getOutput
   * @apiGroup Output
   *
   * @apiParam {Number} outputIndex output index
   *
   * @apiSuccess {Number} outputIndex Output index
   * @apiSuccess {String} outputRoot Output root
   * @apiSuccess {String} stateRoot State root
   * @apiSuccess {String} storageRoot Storage root
   * @apiSuccess {String} lastBlockHash Last block hash in this output
   * @apiSuccess {Number} checkpointBlockHeight Checkpoint height for this output
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
