import pako from 'pako'

// compress tx data to submit L1
export function compress(input: string[]): Buffer {
  const recordsWithCommas = input.join(',')
  const recordsBuffer = Buffer.from(recordsWithCommas)
  return Buffer.from(pako.gzip(recordsBuffer))
}

// decompress indexed batch data
export function decompress(input: Buffer): string[] {
  const decompressed = pako.inflate(input)
  const output: string = Buffer.from(decompressed).toString()
  return output.split(',')
}
