import * as crypto from 'crypto'
import { maxNamespace, minNamespace } from './namespace'

const leafPrefix = 0
const nodePrefix = 1

abstract class Hash {
  abstract write(b: Uint8Array): void
  abstract sum(b?: Uint8Array): Uint8Array
  abstract reset(): void
  abstract size(): number
  abstract blockSize(): number
}

export class SHA256 extends Hash {
  private hash: crypto.Hash
  constructor() {
    super()
    this.hash = crypto.createHash('sha256')
  }
  write(b: Uint8Array) {
    this.hash.write(b)
  }
  sum(b = new Uint8Array([])): Uint8Array {
    const hash = this.hash.digest()
    return new Uint8Array([...b, ...hash])
  }
  reset(): void {
    this.hash = crypto.createHash('sha256')
  }
  size(): number {
    return 32
  }
  blockSize(): number {
    return 64
  }
}

export abstract class Hasher {
  abstract isMaxNamespaceIDIgnored(): boolean
  abstract namespaceSize(): number
  abstract hashLeaf(data: Uint8Array): Uint8Array
  abstract hashNode(leftChild: Uint8Array, rightChild: Uint8Array): Uint8Array
  abstract emptyRoot(): Uint8Array
}

export class MntHasher extends Hasher {
  private precomputedMaxNs: Uint8Array
  private tp: number
  private data: Uint8Array

  constructor(
    private baseHasher: Hash,
    private namespaceLen: number,
    private ignoreMaxNs: boolean
  ) {
    super()
    this.precomputedMaxNs = new Uint8Array(
      new Array(this.namespaceLen).fill(0xff)
    )
    this.tp = 0
    this.data = new Uint8Array()
  }

  isMaxNamespaceIDIgnored(): boolean {
    return this.ignoreMaxNs
  }

  namespaceSize(): number {
    return this.namespaceLen
  }

  hashLeaf(data: Uint8Array): Uint8Array {
    const h = this.baseHasher
    h.reset()

    this.validateLeaf(data)

    const nID = data.slice(0, this.namespaceLen)
    const minMaxNIDs = new Uint8Array([...nID, ...nID])

    const leafPrefixedNData = new Uint8Array([leafPrefix, ...data])
    h.write(leafPrefixedNData)

    return h.sum(minMaxNIDs)
  }

  hashNode(left: Uint8Array, right: Uint8Array): Uint8Array {
    this.validateNodes(left, right)

    const h = this.baseHasher
    h.reset()

    const [leftMinNs, leftMaxNs] = [
      minNamespace(left, this.namespaceLen),
      maxNamespace(left, this.namespaceLen)
    ]
    const [rightMinNs, rightMaxNs] = [
      minNamespace(right, this.namespaceLen),
      maxNamespace(right, this.namespaceLen)
    ]

    const [minNs, MaxNs] = this.computeNsRange(
      leftMinNs,
      leftMaxNs,
      rightMinNs,
      rightMaxNs
    )

    const res = new Uint8Array([...minNs, ...MaxNs])
    const data = new Uint8Array([nodePrefix, ...left, ...right])

    h.write(data)
    return h.sum(res)
  }

  emptyRoot(): Uint8Array {
    this.baseHasher.reset()
    const emptyNs = new Array(this.namespaceLen).fill(0)
    const h = this.baseHasher.sum()
    const digest = [...emptyNs, ...emptyNs, ...h]
    return new Uint8Array(digest)
  }

  private validateLeaf(data: Uint8Array) {
    const nidSize = this.namespaceSize()
    const lenData = data.length
    if (lenData < nidSize) {
      throw Error('invalid leaf len')
    }
  }

  private validateNodes(left: Uint8Array, right: Uint8Array) {
    this.validateNodeFormat(left)
    this.validateNodeFormat(right)
    this.validateSiblingsNamespaceOrder(left, right)
  }

  private validateNodeFormat(node: Uint8Array) {
    const expectNodeLen = this.size()
    const nodeLen = node.length
    if (nodeLen !== expectNodeLen) {
      throw Error('invalid node size')
    }

    const minNID = minNamespace(node, this.namespaceSize())
    const maxNID = maxNamespace(node, this.namespaceSize())
    if (Buffer.from(maxNID).compare(Buffer.from(minNID)) < 0) {
      throw Error('max namespace ID is less than min namespace ID')
    }
  }

  private validateSiblingsNamespaceOrder(left: Uint8Array, right: Uint8Array) {
    const leftMaxNs = maxNamespace(left, this.namespaceSize())
    const rightMinNs = maxNamespace(right, this.namespaceSize())
    if (Buffer.from(rightMinNs).compare(Buffer.from(leftMaxNs)) < 0) {
      throw Error(
        'the max namespace of the left child is greater than the min namespace of right child'
      )
    }
  }

  private size(): number {
    return this.baseHasher.size() + this.namespaceLen * 2
  }

  private computeNsRange(
    leftMinNs: Uint8Array,
    leftMaxNs: Uint8Array,
    rightMinNs: Uint8Array,
    rightMaxNs: Uint8Array
  ): [Uint8Array, Uint8Array] {
    const minNs = leftMinNs
    let maxNs = rightMaxNs
    if (
      this.ignoreMaxNs &&
      Buffer.from(this.precomputedMaxNs).compare(rightMinNs) === 0
    ) {
      maxNs = leftMaxNs
    }
    return [minNs, maxNs]
  }
}
