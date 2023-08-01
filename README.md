# Initia Rollup

Initia Optimistic Rollup Bots.

- Batch Submitter: Submit batch to L1 node
- Output Submitter: Submit output to L1 node
- Challenger: Challenge invalid output
- Bridge Executor: Execute bridge transaction

# How to use

## Batch Submitter

You can run batch submitter and server with following command.

```bash
npm run batch
```

If you use pm2, you can run batch submitter and server with following command.

```bash
pm2 start src/worker/batchSubmitter/pm2.json
```

## Output Submitter

```bash
npm run challenger
```

If you use pm2,
```bash
pm2 start src/worker/outputSubmitter/pm2.json
```

## Challenger

```bash
npm run challenger
```

If you use pm2, 
```bash
pm2 start src/worker/challenger/pm2.json
```

## Bridge Executor

```bash
npm run challenger
```

If you use pm2,
```bash
pm2 start src/worker/bridgeExecutor/pm2.json
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
| EXECUTOR_MNEMONIC         | Mnemonic seed for executor                             | ''                               |
| BATCH_SUBMITTER_MNEMONIC  | Mnemonic seed for submitter                            | ''                               |
| OUTPUT_SUBMITTER_MNEMONIC | Mnemonic seed for output submitter                     | ''                               |
| CHALLENGER_MNEMONIC       | Mnemonic seed for challenger                           | ''                               |

> In Batch Submitter, we use [direnv](https://direnv.net) for managing environment variable for development. See [sample of .envrc](.envrc_sample)
