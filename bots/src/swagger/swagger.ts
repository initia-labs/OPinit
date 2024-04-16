import { SwaggerRouter } from 'koa-swagger-decorator'
import { DepositTxController } from '../controller/executor/DepositTxController'
import { OutputController } from '../controller/executor/OutputController'
import { WithdrawalTxController } from '../controller/executor/WithdrawalTxController'
import { ClaimTxController } from '../controller/executor/ClaimTxController'

const router = new SwaggerRouter({
  spec: {
    info: {
      title: 'Initia VIP API',
      version: 'v1.0'
    }
  },
  swaggerHtmlEndpoint: '/swagger',
  swaggerJsonEndpoint: '/swagger.json'
})

router.swagger()
router
  .applyRoute(DepositTxController)
  .applyRoute(OutputController)
  .applyRoute(WithdrawalTxController)
  .applyRoute(ClaimTxController)

export { router }
