const namespaceVersionSize = 1
export const namespaceIDSize = 28
export const namespaceSize = namespaceVersionSize + namespaceIDSize
export const namespaceVersionZero = 0
const namespaceVersionMax = 255
const namespaceVersionZeroPrefixSize = 18
const namespaceVersionZeroPrefix = new Uint8Array(
  new Array(namespaceVersionZeroPrefixSize).fill(0)
)

export class Namespace {
  constructor(
    private version: number,
    private id: Uint8Array
  ) {
    // validate version
    if (version !== namespaceVersionZero && version !== namespaceVersionMax) {
      throw Error('unsupported namespace version')
    }

    if (id.length !== namespaceIDSize) {
      throw Error('unsupported namespace id lengh')
    }

    // validate id
    if (
      version === namespaceVersionZero &&
      id.slice(0, namespaceVersionZeroPrefix.length) ===
        namespaceVersionZeroPrefix
    ) {
      throw Error('unsupported namespace id with version')
    }
  }

  isTx(): boolean {
    return this.bytes() === txNamespace.bytes()
  }

  isPayForBlob(): boolean {
    return this.bytes() === payForBlobNamespace.bytes()
  }

  bytes(): Uint8Array {
    return new Uint8Array([this.version, ...this.id])
  }
}

export function minNamespace(hash: Uint8Array, size: number): Uint8Array {
  return new Uint8Array([...hash.slice(0, size)])
}

export function maxNamespace(hash: Uint8Array, size: number): Uint8Array {
  return new Uint8Array([...hash.slice(size, size * 2)])
}

export function primaryReservedNamespace(lastByte: number): Namespace {
  const ns = new Namespace(
    namespaceVersionZero,
    new Uint8Array([...new Array(namespaceIDSize - 1).fill(0x00), lastByte])
  )

  return ns
}

const txNamespace = primaryReservedNamespace(0x01)
const payForBlobNamespace = primaryReservedNamespace(0x04)
