import * as pako from 'pako';

// compress tx data to submit L1
export function compressor(input: string[]): Buffer {
  const recordsWithCommas = input.join(',');
  const recordsBuffer = Buffer.from(recordsWithCommas);
  return pako.gzip(recordsBuffer);
}

// decompress indexed batch data
export function decompressor(input: Buffer): string[] {
  const decompressed = pako.inflate(input);
  const output: string = Buffer.from(decompressed).toString();
  return output.split(',');
}
