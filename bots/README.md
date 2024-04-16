# OPinit Bots

Initia Optimistic Rollup Bots.

- Batch Submitter: Submit batch to L1 node
- Output Submitter: Submit output to L1 node
- Challenger: Challenge invalid output
- Bridge Executor: Execute bridge transaction

## How to use

### Prerequisites

- Postgres 14+
- Node.js 16+
- Node LCD/RPC (L1 and L2)

### Step1. Create Bridge

Before running rollup bots, you should create bridge between L1 and L2. If you use `initia.js`, you can create bridge using `MsgCreateBridge` message as follows.

```typescript
import { MsgCreateBridge, BridgeConfig, Duration } from '@initia/initia.js';

const bridgeConfig = new BridgeConfig(
  challenger.key.accAddress,
  outputSubmitter.key.accAddress,
  new BatchInfo(batchSubmitter.accAddress, config.PUBLISH_BATCH_TARGET),
  Duration.fromString(submissionInterval.toString()),
  Duration.fromString(finalizedTime.toString()),
  new Date(),
  this.metadata
);
const msg = new MsgCreateBridge(executor.key.accAddress, bridgeConfig);
```

### Step2. Configuration

You should set `.env` file for each bot in `bots/worker`. To transfer assets between L1 and L2, you should run `executor` and `output submitter` at least.

> In OPinit bots, we use [.dotenv](https://www.npmjs.com/package/dotenv) for managing environment variable for development. If you want to set `.env` by worker, you should name it as `.env.{WORKER_NAME}` and set `WORKER_NAME` in [`executor`, `output`, `batch`, `challenger`].
> For example, if you want to set `.env` for `executor`, you should name it as `.env.executor` and set `WORKER_NAME=executor` in local environment. `.env` files should be located in `OPinit/bots` directory.

- `.env.executor`

| Name                       | Description                                | Default                  |
| -------------------------- | ------------------------------------------ | ------------------------ |
| L1_LCD_URI                 | L1 node LCD URI                            | <http://127.0.0.1:1317>  |
| L1_RPC_URI                 | L1 node RPC URI                            | <http://127.0.0.1:26657> |
| L2_LCD_URI                 | L2 node LCD URI                            | <http://127.0.0.1:1317>  |
| L2_RPC_URI                 | L2 node RPC URI                            | <http://127.0.0.1:26657> |
| L2_GAS_PRICES              | Gas prices for L2 chain                    | '0.15umin'               |
| BRIDGE_ID                  | Bridge ID                                  | ''                       |
| EXECUTOR_PORT              | Executor port                              | 5000                     |
| EXECUTOR_MNEMONIC          | Mnemonic seed for executor                 | ''                       |
| SLACK_WEB_HOOK             | Slack web hook for notification (optional) | ''                       |
| EXECUTOR_L1_MONITOR_HEIGHT | L1 monitor start height (optional)         | 0                        |
| EXECUTOR_L2_MONITOR_HEIGHT | L2 monitor start height (optional)         | 0                        |
| ENABLE_API_ONLY            | Enable API only mode (optional)            | false                    |

> Note that if `EXECUTOR_L1_MONITOR_HEIGHT` and `EXECUTOR_L2_MONITOR_HEIGHT` are not set, `executor` will start monitoring from height stored on `state` table. If you want to start monitoring from specific height, you should set them in `.env.executor` file.

- `.env.output`

| Name                      | Description                                | Default                  |
| ------------------------- | ------------------------------------------ | ------------------------ |
| L1_LCD_URI                | L1 node LCD URI                            | <http://127.0.0.1:1317>  |
| L1_RPC_URI                | L1 node RPC URI                            | <http://127.0.0.1:26657> |
| BRIDGE_ID                 | Bridge ID                                  | ''                       |
| OUTPUT_SUBMITTER_MNEMONIC | Mnemonic seed for output submitter         | ''                       |
| EXECUTOR_URI              | Executor URI                               | <http://localhost:5000>  |
| SLACK_WEB_HOOK            | Slack web hook for notification (optional) | ''                       |

- `.env.batch`

| Name                        | Description                                                  | Default                    |
| --------------------------- | ------------------------------------------------------------ | -------------------------- |
| L1_LCD_URI                  | L1 node LCD URI                                              | <http://127.0.0.1:1317>    |
| L1_RPC_URI                  | L1 node RPC URI                                              | <http://127.0.0.1:26657>   |
| L2_LCD_URI                  | L2 node LCD URI                                              | <http://127.0.0.1:1317>    |
| L2_RPC_URI                  | L2 node RPC URI                                              | <http://127.0.0.1:26657>   |
| BRIDGE_ID                   | Bridge ID                                                    | ''                         |
| BATCH_PORT                  | Batch submitter port                                         | 5001                       |
| BATCH_SUBMITTER_MNEMONIC    | Mnemonic seed for submitter                                  | ''                         |
| SLACK_WEB_HOOK              | Slack web hook for notification (optional)                   | ''                         |
| BATCH_CHAIN_ID              | DA chain's chain-id                                          |                            |
| BATCH_CHAIN_RPC_URI         | DA chain node RPC URI                                        | L1_RPC_URI if target is l1 |
| BATCH_LCD_URI               | DA chain node LCD URI                                        | <http://127.0.0.1:1317>    |
| BATCH_GAS_PRICES            | Gas prices for DA chain                                      |                            |
| BATCH_DENOM                 | Fee denom for DA chain                                       |                            |
| CELESTIA_NAMESPACE_ID       | Celestia namespace id (optional)                             | ''                         |
| PUBLISH_BATCH_TARGET        | Target chain to publish batch (supports: ['l1', 'celestia']) | 'l1'                       |
| ENABLE_API_ONLY             | Enable API only mode (optional)                              | false                      |

- `.env.challenger`

| Name                   | Description                                | Default                  |
| ---------------------- | ------------------------------------------ | ------------------------ |
| L1_LCD_URI             | L1 node LCD URI                            | <http://127.0.0.1:1317>  |
| L1_RPC_URI             | L1 node RPC URI                            | <http://127.0.0.1:26657> |
| L2_LCD_URI             | L2 node LCD URI                            | <http://127.0.0.1:1317>  |
| L2_RPC_URI             | L2 node RPC URI                            | <http://127.0.0.1:26657> |
| BRIDGE_ID              | Bridge ID                                  | ''                       |
| CHALLENGER_MNEMONIC    | Mnemonic seed for challenger               | ''                       |
| DELETE_OUTPUT_PROPOSAL | Enable delete output proposal instantly    | ''                       |
| SLACK_WEB_HOOK         | Slack web hook for notification (optional) | ''                       |

### Step3. Run Bots

- Install dependencies

  ```bash
  npm install
  ```

- Bridge Executor

Bridge executor is a bot that monitor L1, L2 node and execute bridge transaction. It will execute following steps.

1. Configure `.env.executor` file
2. Run executor bot

   ```bash
   npm run executor
   ```

- Output Submitter

Output submitter is the component to store the L2 output root for block finalization.
Output submitter will get the L2 output results from executor and submit it to L1.

> Before running output submitter, you should set `executor` first. It will get the L2 output results from `executor` and submit it to L1.

1. Configure `.env.output` file
2. Run output submitter bot

   ```bash
   npm run output
   ```

- Batch Submitter

Batch submitter is a background process that submits transaction batches to the BatchInbox module of L1 or Celestia.

> **NOTE** To run celestia batch submitter, you have to set [celestia light node](https://docs.celestia.org/nodes/light-node)

1. Configure `.env.batch` file
2. Run batch submitter bot

   ```bash
   npm run batch
   ```

- Challenger

Challenger is an entity capable of deleting invalid output proposals from the output oracle.

> **NOTE** > `challenger` should be independent from `executor` and `output submitter`. It will monitor the output oracle and delete invalid output proposals.

1. Configure `.env.challenger` file
2. Run challenger bot

   ```bash
   npm run challenger
   ```
