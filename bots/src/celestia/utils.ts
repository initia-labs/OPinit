import { sequenceLenBytes, shareInfoBytes, shareSize } from './builder'
import { createCommitment } from './commitment'
import { namespaceSize } from './namespace'
import { config } from '../config'
import { Blob } from '@initia/initia.js'

// constants
const defaultGasPerBlobByte = 8
const defaultTxSizeCostPerByte = 10
const bytesPerBlobInfo = 70
const pfbGasFixedCost = 75000
const firstSparseShareContentSize =
  shareSize - namespaceSize - shareInfoBytes - sequenceLenBytes
const continuationSparseShareContentSize =
  shareSize - namespaceSize - shareInfoBytes
const namespaceIdLength = 28

export function getCelestiaFeeGasLimit(length: number): number {
  // calculate gas
  return Math.floor(defaultEstimateGas([length]) * 1.2)
}

export function createBlob(data: Buffer): {
  blob: Blob;
  commitment: string;
  namespace: string;
} {
  const blob = {
    namespaceId: new Uint8Array([
      ...Buffer.from(config.CELESTIA_NAMESPACE_ID, 'hex')
    ]),
    data: new Uint8Array([...data]),
    shareVersion: 0,
    namespaceVersion: 0,
    shareCommitment: new Uint8Array()
  }

  // generate commitment
  createCommitment(blob)

  const commitment = Buffer.from(blob.shareCommitment).toString('base64')
  const res = new Blob(
    Buffer.from(config.CELESTIA_NAMESPACE_ID, 'hex').toString('base64'),
    Buffer.from(blob.data).toString('base64'),
    0,
    0
  )
  const namespace = Buffer.from([
    blob.namespaceVersion,
    ...blob.namespaceId
  ]).toString('base64')
  return {
    blob: res,
    commitment,
    namespace
  }
}

function defaultEstimateGas(blobSizes: number[]) {
  return estimateGas(
    blobSizes,
    defaultGasPerBlobByte,
    defaultTxSizeCostPerByte
  )
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
  )
}

function gasToConsume(blobSizes: number[], gasPerByte: number): number {
  let totalSharesUsed = 0
  for (const size of blobSizes) {
    totalSharesUsed += sparseSharesNeeded(size)
  }

  return totalSharesUsed * shareSize * gasPerByte
}

function sparseSharesNeeded(sequenceLen: number): number {
  if (sequenceLen === 0) {
    return 0
  }

  if (sequenceLen < firstSparseShareContentSize) {
    return 1
  }

  return (
    1 +
    Math.floor(
      (sequenceLen - firstSparseShareContentSize) /
        continuationSparseShareContentSize
    )
  )
}

export function validateCelestiaConfig() {
  if (config.PUBLISH_BATCH_TARGET !== 'celestia') return
  const length = Buffer.from(config.CELESTIA_NAMESPACE_ID, 'hex').length
  // https://docs.celestia.org/developers/blobstream-proof-queries#namespace
  if (length != namespaceIdLength) {
    throw Error(
      `unsupported namespace id length; expected ${namespaceIdLength}, given ${length} `
    )
  }
}

export function roundUpPowerOfTwo(input: number) {
  let result = 1
  while (result < input) {
    result <<= 1
  }
  return result
}

export function roundDownPowerOfTwo(input: number): number {
  if (input <= 0) {
    throw Error('input must be positive')
  }
  const roundedUp = roundUpPowerOfTwo(input)
  if (roundedUp == input) {
    return roundedUp
  }
  return roundedUp >> 1
}
