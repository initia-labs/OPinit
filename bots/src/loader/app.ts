// import { controllers } from 'controller'
import * as Koa from 'koa';
import * as bodyParser from 'koa-body';
import * as Router from 'koa-router';
import * as cors from '@koa/cors';
import { errorHandler, APIError, ErrorTypes } from 'lib/error';
import { error } from 'lib/response';
import { KoaController, configureRoutes } from 'koa-joi-controllers';
import { createApiDocApp } from './apidoc';
import * as mount from 'koa-mount';

const API_VERSION_PREFIX = '/v1';

const notFoundMiddleware: Koa.Middleware = (ctx) => {
  ctx.status = 404;
};

export async function initApp(controllers: KoaController[]): Promise<Koa> {
  const app = new Koa();

  const apiDocApp = createApiDocApp();

  app.proxy = true;
  app
    .use(cors())
    .use(errorHandler(error))
    .use(async (ctx, next) => {
      await next();

      ctx.set('Cache-Control', 'no-store, no-cache, must-revalidate');
      ctx.set('Pragma', 'no-cache');
      ctx.set('Expires', '0');
    })
    .use(
      bodyParser({
        multipart: true,
        onError: (error) => {
          throw new APIError(
            ErrorTypes.INVALID_REQUEST_ERROR,
            undefined,
            error.message
          );
        }
      })
    );

  const router = new Router();

  router.get('/health', async (ctx) => {
    ctx.status = 200;
    ctx.body = 'OK';
  });

  app.use(router.routes());
  app.use(router.allowedMethods());
  app.use(mount('/apidoc', apiDocApp));
  configureRoutes(app, controllers);
  app.use(notFoundMiddleware);

  return app;
}
