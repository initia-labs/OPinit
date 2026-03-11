# Protobufs

This is the public protocol buffers API for the [OPinit](https://github.com/initia-labs/OPinit).

## npm Package

TypeScript definitions are published to npm as [`@initia/opinit-proto`](https://www.npmjs.com/package/@initia/opinit-proto) on every tagged release (`v*`).

### Installation

```bash
npm install @initia/opinit-proto @bufbuild/protobuf
```

### Usage

```typescript
import { MsgRecordBatchSchema } from "@initia/opinit-proto/opinit/ophost/v1/tx_pb";
import { MsgFinalizeTokenDepositSchema } from "@initia/opinit-proto/opinit/opchild/v1/tx_pb";
```

The package requires `@bufbuild/protobuf` v2 as a peer dependency.
