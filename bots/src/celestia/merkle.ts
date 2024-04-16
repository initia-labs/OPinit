import * as crypto from 'crypto'
import { Hasher, MntHasher } from './hasher'
import { roundDownPowerOfTwo } from './utils'

export class NamespaceMerkleTree {
  private treeHasher: Hasher
  private visit: nodeVisitorFn
  private leaves: Uint8Array[]
  private leafHashes: Uint8Array[]
  private namespaceRanges: Record<string, leafRange>
  private minNID: Uint8Array
  private maxNID: Uint8Array
  private rawRoot: Uint8Array | undefined
  constructor(hash: any, namespaceIDSize: number, ignoreMaxNamespace: boolean) {
    this.treeHasher = new MntHasher(hash, namespaceIDSize, ignoreMaxNamespace)
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    this.visit = (_hash: Uint8Array, _children?: Uint8Array[]) => {} // no op
    this.leaves = []
    this.leafHashes = []
    this.namespaceRanges = {}
    this.minNID = new Uint8Array(new Array(namespaceIDSize).fill(0xff))
    this.maxNID = new Uint8Array(new Array(namespaceIDSize).fill(0x00))
  }
  push(namespacedData: Uint8Array) {
    const nID = this.validateAndExtractNamespace(namespacedData)

    const res = this.treeHasher.hashLeaf(namespacedData)

    this.leaves.push(namespacedData)
    this.leafHashes.push(res)
    this.updateNamespaceRanges()
    this.updateMinMaxID(nID)
    this.rawRoot = undefined
    return undefined
  }
  root(): Uint8Array {
    if (this.rawRoot === undefined) {
      const res = this.computeRoot(0, this.size())
      this.rawRoot = res
    }

    return this.rawRoot
  }

  computeRoot(start: number, end: number): Uint8Array {
    if (start < 0 || start > end || end > this.size()) {
      throw Error('fail to compute root')
    }

    switch (end - start) {
      case 0: {
        const rootHash = this.treeHasher.emptyRoot()
        this.visit(rootHash)
        return rootHash
      }
      case 1: {
        const leafHash = new Uint8Array([...this.leafHashes[start]]) // copy
        this.visit(leafHash, [this.leaves[start]])
        return leafHash
      }
      default: {
        const k = getSplitPoint(end - start)
        const left = this.computeRoot(start, start + k)
        const right = this.computeRoot(start + k, end)
        const hash = this.treeHasher.hashNode(left, right)
        this.visit(hash, [left, right])
        return hash
      }
    }
  }

  private validateAndExtractNamespace(data: Uint8Array): Uint8Array {
    const nidSize = this.namespaceSize()
    if (data.length < nidSize) {
      throw Error('invalid data length')
    }

    const nID = data.slice(0, this.namespaceSize())

    const curSize = this.size()
    if (curSize > 0) {
      if (
        Buffer.from(nID).compare(
          Buffer.from(this.leaves[curSize - 1].slice(0, nidSize))
        ) < 0
      ) {
        throw Error('invalid push order')
      }
    }

    return nID
  }

  private updateNamespaceRanges() {
    if (this.size() > 0) {
      const lastIndex = this.size() - 1
      const lastPushed = this.leaves[lastIndex]

      const lastNsStr = Buffer.from(
        lastPushed.slice(0, this.treeHasher.namespaceSize())
      ).toString()

      const lastRange = this.namespaceRanges[lastNsStr]

      if (lastRange === undefined) {
        this.namespaceRanges[lastNsStr] = {
          start: lastIndex,
          end: lastIndex + 1
        }
      } else {
        this.namespaceRanges[lastNsStr] = {
          start: lastRange.start,
          end: lastRange.end + 1
        }
      }
    }
  }

  private updateMinMaxID(id: Uint8Array) {
    if (Buffer.from(id).compare(this.minNID) < 0) {
      this.minNID = id
    }

    if (Buffer.from(this.maxNID).compare(id) < 0) {
      this.maxNID = id
    }
  }

  private size() {
    return this.leaves.length
  }

  private namespaceSize(): number {
    return this.treeHasher.namespaceSize()
  }
}

type nodeVisitorFn = (hash: Uint8Array, children?: Uint8Array[]) => void

interface leafRange {
  start: number;
  end: number;
}

function getSplitPoint(length: number): number {
  if (length < 1) {
    throw Error('Trying to split a tree with size < 1')
  }

  let k = 1
  while (k < length) {
    k <<= 1
  }

  k >>= 1
  if (k === length) {
    k >>= 1
  }

  return k
}

export function hashFromByteSlices(items: Uint8Array[]): Uint8Array {
  switch (items.length) {
    case 0:
      return new Uint8Array([...crypto.createHash('sha256').digest()]) // emptyHash
    case 1: {
      const hash = crypto.createHash('sha256')
      hash.write(Buffer.from([0 /* leaf prefix */, ...items[0]]))
      return new Uint8Array([...hash.digest()])
    }
    default: {
      const k = getSplitPoint(items.length)
      const left = hashFromByteSlices(items.slice(0, k))
      const right = hashFromByteSlices(items.slice(k))
      const hash = crypto.createHash('sha256')
      hash.write(Buffer.from([1 /* inner prefix */, ...left, ...right]))
      return new Uint8Array([...hash.digest()])
    }
  }
}

export function merkleMountainRangeSizes(
  totalSize: number,
  maxTreeSize: number
): number[] {
  let treeSizes: number[] = []

  while (totalSize !== 0) {
    if (totalSize >= maxTreeSize) {
      treeSizes = [...treeSizes, maxTreeSize]
      totalSize -= maxTreeSize
    } else {
      const treeSize = roundDownPowerOfTwo(totalSize)

      treeSizes = [...treeSizes, treeSize]
      totalSize -= treeSize
    }
  }
  return treeSizes
}
