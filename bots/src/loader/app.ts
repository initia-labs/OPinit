import path from 'path'
import Koa from 'koa'
import bodyParser from 'koa-body'
import Router from 'koa-router'
import cors from '@koa/cors'
import morgan from 'koa-morgan'
// import  helmet from 'koa-helmet'
import serve from 'koa-static'
import mount from 'koa-mount'
import { APIError, ErrorTypes, errorHandler } from '../lib/error'
import { error } from '../lib/response'
import { KoaController, configureRoutes } from 'koa-joi-controllers'
import { router as swaggerRouter } from '../swagger/swagger'

const notFoundMiddleware: Koa.Middleware = (ctx) => {
  ctx.status = 404
}

function getRootApp(): Koa {
  // root app only contains the health check route
  const app = new Koa()
  const router = new Router()

  router.get('/health', async (ctx) => {
    ctx.status = 200
    ctx.body = 'OK'
  })

  app.use(router.routes())
  app.use(router.allowedMethods())

  return app
}

function createApiDocApp(): Koa {
  // static
  const app = new Koa()

  app
    .use(
      serve(path.resolve(__dirname, '..', 'static'), {
        maxage: 86400 * 1000
      })
    )
    .use(notFoundMiddleware)

  return app
}

async function createAPIApp(controllers: KoaController[]): Promise<Koa> {
  const app = new Koa()

  app
    .use(errorHandler(error))
    .use(async (ctx, next) => {
      await next()

      ctx.set('Cache-Control', 'no-store, no-cache, must-revalidate')
      ctx.set('Pragma', 'no-cache')
      ctx.set('Expires', '0')
    })
    .use(
      bodyParser({
        formLimit: '512kb',
        jsonLimit: '512kb',
        textLimit: '512kb',
        multipart: true,
        onError: (error) => {
          throw new APIError(
            ErrorTypes.INVALID_REQUEST_ERROR,
            '',
            error.message,
            error
          )
        }
      })
    )

  configureRoutes(app, controllers)
  app.use(notFoundMiddleware)
  return app
}

export async function initApp(controllers: KoaController[]): Promise<Koa> {
  const app = getRootApp()

  app.proxy = true

  const apiDocApp = createApiDocApp()
  const apiApp = await createAPIApp(controllers)

  app
    .use(morgan('common'))
    // .use(
    //   helmet({
    //     contentSecurityPolicy: {
    //       directives: {
    //         defaultSrc: [`'self'`, 'http:'],
    //         scriptSrc: [
    //           `'self'`,
    //           `'unsafe-inline'`,
    //           `'unsafe-eval'`,
    //           'http:',
    //           'cdnjs.cloudflare.com',
    //           'unpkg.com',
    //         ],
    //         fontSrc: [`'self'`, 'http:', 'https:', 'data:'],
    //         objectSrc: [`'none'`],
    //         imgSrc: [`'self'`, 'http:', 'data:', 'validator.swagger.io'],
    //         styleSrc: [`'self'`, 'http:', 'https:', `'unsafe-inline'`],
    //         blockAllMixedContent: [],
    //       },
    //     },
    //     crossOriginEmbedderPolicy: false,
    //   })
    // )
    .use(cors())
    .use(swaggerRouter.routes())
    .use(swaggerRouter.allowedMethods())
    .use(mount('/apidoc', apiDocApp))
    .use(mount('/', apiApp))

  return app
}
