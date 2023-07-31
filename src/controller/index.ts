import { KoaController } from 'koa-joi-controllers'
import BatchController from './BatchController'

export const controllers = [
  BatchController
]
  .map((prototype) => {
    const controller = new prototype()
    controller.routes = controller.routes.filter((route) => {
      return true
    })

    return controller
  })
  .filter(Boolean) as KoaController[]
export default controllers