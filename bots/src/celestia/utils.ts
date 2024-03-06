import Axios from 'axios';
import { sequenceLenBytes, shareInfoBytes, shareSize } from './builder';
import { createCommitment } from './commitment';
import { namespaceSize } from './namespace';
import { config } from 'config';
import { delay } from 'bluebird';

// constants
const befaultGasPerBlobByte = 8;
const defaultTxSizeCostPerByte = 10;
const bytesPerBlobInfo = 70;
const pfbGasFixedCost = 75000;
const firstSparseShareContentSize =
  shareSize - namespaceSize - shareInfoBytes - sequenceLenBytes;
const continuationSparseShareContentSize =
  shareSize - namespaceSize - shareInfoBytes;

export async function safeSubmitPayForBlob(data: Buffer): Promise<string[]> {
  const maxBytes = 500000; // 500kb

  const batchInfos: string[] = [];

  while (data.length !== 0) {
    let subData: Buffer;
    if (data.length > maxBytes) {
      subData = data.slice(0, maxBytes);
      data = data.slice(maxBytes);
    } else {
      subData = data;
      data = Buffer.from([]);
    }

    const batchInfo = await submitPayForBlob(subData);
    batchInfos.push(batchInfo);
    await delay(1000); // break for each tx ended
  }

  return batchInfos;
}

async function submitPayForBlob(data: Buffer): Promise<string> {
  const lightNodeRpc = Axios.create({
    baseURL: config.CELESTIA_LIGHT_NODE_RPC_URI,
    headers: {
      Authorization: `Bearer ${config.CELESTIA_TOKEN_AUTH}`
    }
  });

  const blob = {
    namespaceId: new Uint8Array([
      ...Buffer.from(config.CELESTIA_NAMESPACE_ID, 'hex')
    ]),
    data: new Uint8Array([...data]),
    shareVersion: 0,
    namespaceVersion: 0,
    shareCommitment: new Uint8Array()
  };

  // generate commitment
  createCommitment(blob);

  // calculate gas
  const gaslimit = Math.floor(defaultEstimateGas([blob.data.length]) * 1.2);
  const fee = Math.floor(gaslimit * config.CELESTIA_GAS_PRICE).toString();

  const request = {
    id: 1,
    jsonrpc: '2.0',
    method: 'state.SubmitPayForBlob',
    params: [
      fee,
      gaslimit,
      [
        {
          namespace: Buffer.from([
            blob.namespaceVersion,
            ...blob.namespaceId
          ]).toString('base64'),
          data: Buffer.from(blob.data).toString('base64'),
          share_version: blob.shareVersion,
          commitment: Buffer.from(blob.shareCommitment).toString('base64')
        }
      ]
    ]
  };

  const response = await lightNodeRpc.post('', request);

  // error handle
  if (response?.data === undefined || response.data?.result === undefined) {
    const timeoutError = 'timed out waiting for tx to be included in a block';
    const mempoolError = 'tx already in mempool';
    // if got timeout error, retry
    if (
      response?.data?.error &&
      (errorInclude(response, timeoutError) ||
        errorInclude(response, mempoolError))
    ) {
      await delay(1000);
      return submitPayForBlob(data);
    }

    let reason: any = '';

    // in case response.data is undefined
    if (response?.data === undefined) {
      reason = response;
    }

    if (response.data?.result === undefined) {
      throw Error(
        `Failed to SubmitPayForBlob. Reason: ${
          response.data?.error
            ? JSON.stringify(response.data.error)
            : JSON.stringify(response.data)
        }`
      );
    }

    throw Error(
      `Failed to SubmitPayForBlob. Reason: ${JSON.stringify(reason)}`
    );
  }

  const height = response.data.result.height;
  const commitment = Buffer.from(blob.shareCommitment).toString('base64');
  return `${height}::${commitment}`;
}

function errorInclude(response: any, message: string): boolean {
  return (
    response?.data?.error &&
    JSON.stringify(response.data.error).indexOf(message) !== -1
  );
}

function defaultEstimateGas(blobSizes: number[]) {
  return estimateGas(
    blobSizes,
    befaultGasPerBlobByte,
    defaultTxSizeCostPerByte
  );
}

function estimateGas(
  blobSizes: number[],
  gasPerByte: number,
  txSizeCost: number
): number {
  return (
    gasToConsume(blobSizes, gasPerByte) +
    txSizeCost * bytesPerBlobInfo * blobSizes.length +
    pfbGasFixedCost
  );
}

function gasToConsume(blobSizes: number[], gasPerByte: number): number {
  let totalSharesUsed = 0;
  for (const size of blobSizes) {
    totalSharesUsed += sparseSharesNeeded(size);
  }

  return totalSharesUsed * shareSize * gasPerByte;
}

function sparseSharesNeeded(sequenceLen: number): number {
  if (sequenceLen === 0) {
    return 0;
  }

  if (sequenceLen < firstSparseShareContentSize) {
    return 1;
  }

  // let sharesNeeded = 1;
  // let bytesAvailable = firstSparseShareContentSize;
  // while (bytesAvailable < sequenceLen) {
  //   bytesAvailable += continuationSparseShareContentSize;
  //   sharesNeeded++;
  // }

  // return sharesNeeded;

  return (
    1 +
    Math.floor(
      (sequenceLen - firstSparseShareContentSize) /
        continuationSparseShareContentSize
    )
  );
}

export function validateCelestiaConfig() {
  if (config.PUBLISH_BATCH_TARGET !== 'celestia') return;
  if (config.CELESTIA_NAMESPACE_ID === '') {
    throw Error(
      'Please check your configuration; CELESTIA_NAMESPACE_ID is needed but not given.'
    );
  }

  if (config.CELESTIA_RPC_URI === undefined) {
    throw Error(
      'Please check your configuration; CELESTIA_RPC_URI is needed but not given.'
    );
  }
}

export function roundUpPowerOfTwo(input: number) {
  let result = 1;
  while (result < input) {
    result <<= 1;
  }
  return result;
}

export function roundDownPowerOfTwo(input: number): number {
  if (input <= 0) {
    throw Error('input must be positive');
  }
  const roundedUp = roundUpPowerOfTwo(input);
  if (roundedUp == input) {
    return roundedUp;
  }
  return roundedUp >> 1;
}
