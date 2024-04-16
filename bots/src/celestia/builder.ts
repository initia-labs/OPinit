import { Namespace } from './namespace'
import { Share } from './share'

// builder constants
export const sequenceLenBytes = 4
const compactShareReservedBytes = 4
export const shareInfoBytes = 1
export const shareSize = 512
const maxShareVersion = 127

export class Builder {
  private isCompactShare: boolean
  private rawShareData: Uint8Array
  constructor(
    private namespace: Namespace,
    private shareVersion: number,
    private isFirstShare: boolean
  ) {
    this.isCompactShare = namespace.isTx() || namespace.isPayForBlob()
    this.rawShareData = new Uint8Array()
  }

  init() {
    if (this.isCompactShare) {
      this.prepareCompactShare()
    } else {
      this.prepareSparseShare()
    }
  }

  prepareCompactShare() {
    let shareData: number[] = []
    const infoByte = newInfoByte(this.shareVersion, this.isFirstShare)
    const placeholderSequenceLen = new Array(sequenceLenBytes).fill(0)
    const placeholderReservedBytes = new Array(compactShareReservedBytes).fill(
      0
    )

    shareData = shareData.concat([...this.namespace.bytes()])
    shareData.push(infoByte)

    if (this.isCompactShare) {
      shareData = shareData.concat(placeholderSequenceLen)
    }

    shareData = shareData.concat(placeholderReservedBytes)

    this.rawShareData = new Uint8Array(shareData)
  }

  prepareSparseShare() {
    let shareData: number[] = []
    const infoByte = newInfoByte(this.shareVersion, this.isFirstShare)
    const placeholderSequenceLen = new Array(sequenceLenBytes).fill(0)

    shareData = shareData.concat([...this.namespace.bytes()])
    shareData.push(infoByte)

    if (this.isCompactShare) {
      shareData = shareData.concat(placeholderSequenceLen)
    }

    this.rawShareData = new Uint8Array(shareData)
  }

  writeSequenceLen(sequenceLen: number) {
    if (!this.isFirstShare) {
      throw Error('not the first share')
    }

    const sequenceLenBuf = [
      (sequenceLen >> 24) % 256,
      (sequenceLen >> 16) % 256,
      (sequenceLen >> 8) % 256,
      sequenceLen % 256
    ]

    this.rawShareData = new Uint8Array([
      ...this.rawShareData,
      ...sequenceLenBuf
    ])
  }

  addData(rawData: Uint8Array): Uint8Array | undefined {
    const pendingLeft = shareSize - this.rawShareData.length

    if (rawData.length <= pendingLeft) {
      this.rawShareData = new Uint8Array([...this.rawShareData, ...rawData])
      return
    }

    const chunk = rawData.slice(0, pendingLeft)
    this.rawShareData = new Uint8Array([...this.rawShareData, ...chunk])

    return rawData.slice(pendingLeft)
  }

  zeroPadIfNecessary(): number {
    const share = this.rawShareData
    const oldLen = share.length

    if (oldLen >= shareSize) {
      this.rawShareData = share
      return 0
    }

    const missingBytes = shareSize - oldLen
    this.rawShareData = new Uint8Array([
      ...this.rawShareData,
      ...new Array(missingBytes).fill(0)
    ])

    return missingBytes
  }

  build(): Share {
    return new Share(this.rawShareData)
  }
}

function newInfoByte(version: number, isSequenceStart: boolean): InfoByte {
  if (version > maxShareVersion) {
    throw Error('version must be less than or equal to max share version')
  }

  const prefix = version << 1
  if (isSequenceStart) {
    return prefix + 1
  }
  return prefix
}

type InfoByte = number
