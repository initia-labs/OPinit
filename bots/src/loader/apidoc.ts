import Koa from 'koa';
import * as serve from 'koa-static';
import * as path from 'path';

const notFoundMiddleware: Koa.Middleware = (ctx) => {
  ctx.status = 404;
};

export function createApiDocApp(): Koa {
  // static
  const app = new Koa();

  app
    .use(
      serve(path.resolve(__dirname, '../../', 'static'), {
        maxage: 86400 * 1000
      })
    )
    .use(notFoundMiddleware);

  return app;
}
