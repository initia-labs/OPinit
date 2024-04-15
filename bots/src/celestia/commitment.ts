import { SHA256 } from './hasher'
import {
  NamespaceMerkleTree,
  hashFromByteSlices,
  merkleMountainRangeSizes
} from './merkle'
import { Namespace, namespaceSize } from './namespace'
import { Blob, Share, SparseShareSplitter } from './share'
import { roundUpPowerOfTwo } from './utils'

const SubtreeRootThreshold = 64

export function createCommitment(blob: Blob): Uint8Array {
  const shares = splitBlobs([blob])

  const subTreeWidth_ = subTreeWidth(shares.length, SubtreeRootThreshold)
  const treeSizes = merkleMountainRangeSizes(shares.length, subTreeWidth_)

  const leafSets: Uint8Array[][] = []
  let cursor = 0
  for (const treeSize of treeSizes) {
    leafSets.push(
      shares
        .slice(cursor, cursor + treeSize)
        .map((share) => new Uint8Array(share.data))
    )
    cursor += treeSize
  }

  const subTreeRoots: Uint8Array[] = []
  for (const set of leafSets) {
    const tree = new NamespaceMerkleTree(new SHA256(), namespaceSize, true)
    for (const leaf of set) {
      const ns = new Namespace(blob.namespaceVersion, blob.namespaceId)
      const nsLeaf = new Uint8Array([...ns.bytes(), ...leaf])
      tree.push(nsLeaf)
    }

    const root = tree.root()
    subTreeRoots.push(root)
  }

  blob.shareCommitment = hashFromByteSlices(subTreeRoots)
  return blob.shareCommitment
}

function splitBlobs(blobs: Blob[]): Share[] {
  const writer = new SparseShareSplitter()
  for (const blob of blobs) {
    writer.write(blob)
  }

  return writer.export()
}

function subTreeWidth(
  shareCount: number,
  subtreeRootThreshold: number
): number {
  let s = Math.floor(shareCount / subtreeRootThreshold)

  if (shareCount % subtreeRootThreshold !== 0) {
    s++
  }

  s = roundUpPowerOfTwo(s)

  return Math.min(s, blobMinSquareSize(shareCount))
}

function blobMinSquareSize(shareCount: number): number {
  return roundUpPowerOfTwo(Math.ceil(Math.sqrt(shareCount)))
}
