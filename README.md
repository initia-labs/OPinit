# Initia Rollup

Initia Optimistic Rollup Bots. You can check [Minitia](https://github.com/initia-labs/minitia) spec for more information.
- Batch Submitter: Submit batch to L1 node
- Output Submitter: Submit output to L1 node
- Challenger: Challenge invalid output
- Bridge Executor: Execute bridge transaction

# How to use

## Setup L2

Initializes the L2 id and op-bridge/output contracts.
You should set `submissionInterval`, `finalizedTime` and `l2StartBlockHeight` before initializing.

```bash
export SUB_INTV=10
export FIN_TIME=10
export L2_HEIGHT=1
npm run l2setup
```

## Bridge Executor

Bridge executor is a bot that monitor L1, L2 node and execute bridge transaction. It will execute following steps.

1. Publish L2 ID to L1
    - L2 ID should be published under executor account
2. Initialize bridge contract on L1 with L2 ID
    - Execute `initialize<L2ID>` entry function in `bridge.move`
3. Run executor bot
    - Execute L1, L2 monitor in bridge executor
        ```bash
        npm run executor
        ```
    - If you use pm2, you can run executor with following command.
        ```bash
        pm2 start executor.json
        ```
4. Register coin to bridge store and prepare deposit store
    - Execute `register_token<L2ID, CoinType>`
5. Now you can deposit after registering coin is done

## Batch Submitter

Batch submitter is a background process that submits transaction batches to the BatchInbox module of L1.
You can run with following command.

```bash
npm run batch
```

If you use pm2,

```bash
pm2 start batch.json
```

## Output Submitter

Output submitter is the component to store the L2 output root for block finalization.
Output submitter will get the L2 output results from executor and submit it to L1 using `propose_l2_output<L2ID>` in `output.move`.

```bash
npm run output
```

If you use pm2,
```bash
pm2 start output.json
```

## Challenger

Challenger is an entity capable of deleting invalid output proposals from the output oracle.


```bash
npm run challenger
```

If you use pm2, 
```bash
pm2 start challenger.json
```

# Configuration

| Name                      | Description                                            | Default                          |
| ------------------------- | ------------------------------------------------------ | -------------------------------- |
| L1_LCD_URI                | L1 node LCD URI                                        | https://stone-rest.initia.tech'  |
| L1_RPC_URI                | L1 node RPC URI                                        | https://stone-rpc.initia.tech'   |
| L2_LCD_URI                | L2 node LCD URI                                        | http://localhost:1317            |
| L2_RPC_URI                | L2 node RPC URI                                        | http://localhost:26657           |
| L2ID                      | L2ID                                                   | ''                               |
| BATCH_PORT                | Batch submitter port                                   | 3000                             |
| EXECUTOR_PORT             | Executor port                                          | 3001                             |
| EXECUTOR_URI              | Executor URI (for output submitter)                    | http://localhost:3000            |
| EXECUTOR_MNEMONIC         | Mnemonic seed for executor                             | ''                               |
| BATCH_SUBMITTER_MNEMONIC  | Mnemonic seed for submitter                            | ''                               |
| OUTPUT_SUBMITTER_MNEMONIC | Mnemonic seed for output submitter                     | ''                               |
| CHALLENGER_MNEMONIC       | Mnemonic seed for challenger                           | ''                               |


> In Batch Submitter, we use [direnv](https://direnv.net) for managing environment variable for development. See [sample of .envrc](.envrc_sample)

# Test

Docker and docker-compose are required to run integration test.

```bash
npm run test:integration
```

If you want to reset docker container, run following command.

```bash
./docker-compose-reset
```