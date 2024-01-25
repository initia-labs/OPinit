# OPinit Bots

Initia Optimistic Rollup Bots.

- Batch Submitter: Submit batch to L1 node
- Output Submitter: Submit output to L1 node
- Challenger: Challenge invalid output
- Bridge Executor: Execute bridge transaction

## How to use

## Create Bridge

Before running rollup bots, you should create bridge between L1 and L2. If you use `initia.js`, you can create bridge using `MsgCreateBridge` message as follows.

```typescript 
import { MsgCreateBridge, BridgeConfig, Duration } from '@initia/initia.js';

const bridgeConfig = new BridgeConfig(
    challenger.key.accAddress,
    outputSubmitter.key.accAddress,
    Duration.fromString(submissionInterval.toString()),
    Duration.fromString(finalizedTime.toString()),
    new Date(),
    this.metadata
);
const msg = new MsgCreateBridge(executor.key.accAddress, bridgeConfig);
```

## Configuration

- `.env.executor`

| Name                      | Description                                            | Default                          |
| ------------------------- | ------------------------------------------------------ | -------------------------------- |
| L1_LCD_URI                | L1 node LCD URI                                        | <http://127.0.0.1:1317>          |
| L1_RPC_URI                | L1 node RPC URI                                        | <http://127.0.0.1:26657>         |
| L2_LCD_URI                | L2 node LCD URI                                        | <http://127.0.0.1:1317>          |
| L2_RPC_URI                | L2 node RPC URI                                        | <http://127.0.0.1:26657>         |
| BRIDGE_ID                 | Bridge ID                                              | ''                               |
| EXECUTOR_PORT             | Executor port                                          | 5000                             |
| EXECUTOR_MNEMONIC         | Mnemonic seed for executor                             | ''                               |
| SLACK_WEB_HOOK            | Slack web hook for notification (optional)             | ''                               |

- `.env.output`

| Name                      | Description                                            | Default                          |
| ------------------------- | ------------------------------------------------------ | -------------------------------- |
| L1_LCD_URI                | L1 node LCD URI                                        | <http://127.0.0.1:1317>          |
| L1_RPC_URI                | L1 node RPC URI                                        | <http://127.0.0.1:26657>         |
| BRIDGE_ID                 | Bridge ID                                              | ''                               |
| OUTPUT_SUBMITTER_MNEMONIC | Mnemonic seed for output submitter                     | ''                               |
| SLACK_WEB_HOOK            | Slack web hook for notification (optional)             | ''                               |

- `.env.batch`

| Name                      | Description                                            | Default                          |
| ------------------------- | ------------------------------------------------------ | -------------------------------- |
| L1_LCD_URI                | L1 node LCD URI                                        | <http://127.0.0.1:1317>          |
| L1_RPC_URI                | L1 node RPC URI                                        | <http://127.0.0.1:26657>         |
| L2_LCD_URI                | L2 node LCD URI                                        | <http://127.0.0.1:1317>          |
| L2_RPC_URI                | L2 node RPC URI                                        | <http://127.0.0.1:26657>         |
| BRIDGE_ID                 | Bridge ID                                              | ''                               |
| BATCH_PORT                | Batch submitter port                                   | 5001                             |
| BATCH_SUBMITTER_MNEMONIC  | Mnemonic seed for submitter                            | ''                               |
| SLACK_WEB_HOOK            | Slack web hook for notification (optional)             | ''                               |

- `.env.challenger`

| Name                      | Description                                            | Default                          |
| ------------------------- | ------------------------------------------------------ | -------------------------------- |
| L1_LCD_URI                | L1 node LCD URI                                        | <http://127.0.0.1:1317>          |
| L1_RPC_URI                | L1 node RPC URI                                        | <http://127.0.0.1:26657>         |
| L2_LCD_URI                | L2 node LCD URI                                        | <http://127.0.0.1:1317>          |
| L2_RPC_URI                | L2 node RPC URI                                        | <http://127.0.0.1:26657>         |
| BRIDGE_ID                 | Bridge ID                                              | ''                               |
| CHALLENGER_MNEMONIC       | Mnemonic seed for challenger                           | ''                               |
| DELETE_OUTPUT_PROPOSAL    | Enable delete output proposal instantly                | ''                               |
| SLACK_WEB_HOOK            | Slack web hook for notification (optional)             | ''                               |


> In OPinit bots, we use [.dotenv](https://www.npmjs.com/package/dotenv) for managing environment variable for development. If you want to set `.env` by worker, you should name it as `.env.{WORKER_NAME}` and set `WORKER_NAME` in [`executor`, `output`, `batch`, `challenger`]. 
For example, if you want to set `.env` for `executor`, you should name it as `.env.executor` and set `WORKER_NAME=executor` in local environment.

## Bridge Executor

Bridge executor is a bot that monitor L1, L2 node and execute bridge transaction. It will execute following steps.

1. Set bridge executor mnemonic on `.env`.
    ```bash
    export EXECUTOR_MNEMONIC="..."
    ```
2. Run executor bot
    ```bash
    npm run executor
    ```

## Batch Submitter

Batch submitter is a background process that submits transaction batches to the BatchInbox module of L1.

1. Set batch submitter mnemonic on `.env`.
    ```bash
    export BATCH_SUBMITTER_MNEMONIC="..."
    ```
2. Run batch submitter bot
    ```bash
    npm run batch
    ```

## Output Submitter

Output submitter is the component to store the L2 output root for block finalization.
Output submitter will get the L2 output results from executor and submit it to L1.

1. Set output submitter mnemonic on `.env`.
    ```bash
    export OUTPUT_SUBMITTER_MNEMONIC="..."
    ```
2. Run output submitter bot
    ```bash
    npm run output
    ```

## Challenger

Challenger is an entity capable of deleting invalid output proposals from the output oracle.

1. Set challenger mnemonic on `.env`.
    ```bash
    export CHALLENGER_MNEMONIC="..."
    ```
2. Run challenger bot
    ```bash
    npm run challenger
    ```