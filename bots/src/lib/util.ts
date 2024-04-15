import { SHA3 } from 'sha3'

export function sha3_256(value: Buffer | string | number): Buffer {
  return new SHA3(256).update(toBuffer(value)).digest()
}

function toBuffer(value: any): Buffer {
  if (!Buffer.isBuffer(value)) {
    if (Array.isArray(value)) {
      value = Buffer.from(value)
    } else if (typeof value === 'string') {
      if (isHexString(value)) {
        value = Buffer.from(padToEven(stripHexPrefix(value)), 'hex')
      } else {
        value = Buffer.from(value)
      }
    } else if (typeof value === 'number') {
      value = numberToBuffer(value)
    } else if (value === null || value === undefined) {
      value = Buffer.allocUnsafe(0)
    } else if (value.toArray) {
      // converts a BN to a Buffer
      value = Buffer.from(value.toArray())
    } else {
      throw new Error('invalid type')
    }
  }

  return value
}

function isHexString(value: string, length?: number): boolean {
  if (!value.match(/^0x[0-9A-Fa-f]*$/)) {
    return false
  }

  if (length && value.length !== 2 + 2 * length) {
    return false
  }

  return true
}

function padToEven(value: string): string {
  if (value.length % 2) {
    value = `0${value}`
  }
  return value
}

function stripHexPrefix(value: string): string {
  return isHexPrefixed(value) ? value.slice(2) : value
}

function isHexPrefixed(value: string): boolean {
  return value.slice(0, 2) === '0x'
}

function numberToBuffer(i: number): Buffer {
  return Buffer.from(padToEven(numberToHexString(i).slice(2)), 'hex')
}

function numberToHexString(i: number): string {
  return `0x${i.toString(16)}`
}
