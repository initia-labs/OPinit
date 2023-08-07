import {
  KoaController,
  Validate,
  Get,
  Controller,
  Validator
} from 'koa-joi-controllers';
import { ErrorCodes } from 'lib/error';
import { success } from 'lib/response';
import { getOutput, getLatestOutput, getAllOutputs, getOutputByHeight } from 'service';

const Joi = Validator.Joi;

@Controller('')
export class OutputController extends KoaController {
  /**
   * @api {get} /output/latest Get all output entity
   * @apiName getAllOutputs
   * @apiGroup Output
   *
   * @apiParam {Number} [offset] Use next property from previous result for pagination
   * @apiParam {Number=10,20,100} [limit=20] Size of page
   *
   * @apiSuccess {Object[]} outputs Output list
   */
  @Get('/output')
  @Validate({
    query: {
      limit: Joi.number()
        .default(20)
        .valid(10, 20, 100)
        .description('Items per page'),
      offset: Joi.alternatives(Joi.number(), Joi.string()).description('Offset')
    },
    failure: ErrorCodes.INVALID_REQUEST_ERROR
  })
  async get(ctx): Promise<void> {
    success(ctx, await getAllOutputs(ctx.query));
  }

  /**
   * @api {get} /output/latest Get latest output entity
   * @apiName getLatestOutput
   * @apiGroup Output
   *
   * @apiSuccess {Object[]} outputs Output list
   */
  @Get('/output/latest')
  async getLatestOutput(ctx): Promise<void> {
    success(ctx, await getLatestOutput());
  }

  /**
   * @api {get} /output/:output_index Get output entity by output index
   * @apiName getOutput
   * @apiGroup Output
   *
   * @apiParam {Number} outputIndex output index
   *
   * @apiSuccess {Object} output Output entity
   */
  @Get('/output/:output_index')
  @Validate({
    params: {
      output_index: Joi.number().description('Output index')
    },
    failure: ErrorCodes.INVALID_REQUEST_ERROR
  })
  async getOutput(ctx): Promise<void> {
    success(ctx, await getOutput(ctx.params.output_index));
  }

  /**
   * @api {get} /output/height/:height Get output entity by checkpoint height
   * @apiName getOutputByHeight
   * @apiGroup Output
   *
   * @apiParam {Number} height checkpoint height
   *
   * @apiSuccess {Object} output Output entity
   */
  @Get('/output/height/:height')
  @Validate({
    params: {
      height: Joi.number().description('height')
    },
    failure: ErrorCodes.INVALID_REQUEST_ERROR
  })
  async getOutputByHeight(ctx): Promise<void> {
    success(ctx, await getOutputByHeight(ctx.params.height));
  }
}
