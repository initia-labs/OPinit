import { Builder } from './builder'
import { Namespace } from './namespace'

const supportedShareVersions = [0]

export class Share {
  constructor(public data: Uint8Array) {}
}

export class SparseShareSplitter {
  private shares: Share[]
  constructor(shares?: Share[]) {
    if (!shares) {
      this.shares = []
    } else {
      this.shares = shares
    }
  }

  write(blob: Blob) {
    if (
      supportedShareVersions.find((v) => v === blob.shareVersion) === undefined
    ) {
      throw Error('unsupported share version')
    }

    let rawData: Uint8Array | undefined = blob.data
    const ns = new Namespace(blob.namespaceVersion, blob.namespaceId)

    let b = new Builder(ns, blob.shareVersion, true)
    b.init()

    b.writeSequenceLen(rawData.length)

    while (rawData !== undefined) {
      const rawDataLeftOver = b.addData(rawData)
      if (rawDataLeftOver === undefined) {
        b.zeroPadIfNecessary()
      }

      const share = b.build()

      this.shares.push(share)

      b = new Builder(ns, blob.shareVersion, false)
      b.init()

      rawData = rawDataLeftOver
    }
  }

  export(): Share[] {
    return this.shares
  }
}

export interface Blob {
  namespaceId: Uint8Array;
  data: Uint8Array;
  shareVersion: number;
  namespaceVersion: number;
  shareCommitment: Uint8Array;
}
