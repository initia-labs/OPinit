import { compress, decompress } from './compressor'

describe('Test compress and decompress functions', () => {
  it('should correctly compress and decompress records', () => {
    const records = [
      Buffer.from('Hello').toString('base64'),
      Buffer.from('World').toString('base64')
    ]

    const compressed = compress(records)
    const decompressed = decompress(compressed)
    expect(decompressed).toEqual(records)

    const decompressedStrs = decompressed.map((buffer) => buffer.toString())
    const originalStrs = records.map((buffer) => buffer.toString())
    expect(decompressedStrs).toEqual(originalStrs)
  })
})
