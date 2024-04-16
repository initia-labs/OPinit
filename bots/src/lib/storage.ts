import { MerkleTree } from 'merkletreejs'
import { sha3_256 } from './util'
import { WithdrawalTx } from './types'
import { AccAddress } from '@initia/initia.js'

function convertHexToBase64(hex: string): string {
  return Buffer.from(hex, 'hex').toString('base64')
}

export class WithdrawStorage {
  private tree: MerkleTree

  constructor(txs: Array<WithdrawalTx>) {
    const leaves = txs.map((tx) => {
      const bridge_id_buf = Buffer.alloc(8)
      bridge_id_buf.writeBigInt64BE(tx.bridge_id)

      const sequence_buf = Buffer.alloc(8)
      sequence_buf.writeBigInt64BE(tx.sequence)

      const amount_buf = Buffer.alloc(8)
      amount_buf.writeBigInt64BE(tx.amount)

      return sha3_256(
        Buffer.concat([
          bridge_id_buf,
          sequence_buf,
          AccAddress.toBuffer(tx.sender),
          AccAddress.toBuffer(tx.receiver),
          Buffer.from(tx.l1_denom, 'utf8'),
          amount_buf
        ])
      )
    })

    this.tree = new MerkleTree(leaves, sha3_256, { sort: true })
  }

  public getMerkleRoot(): string {
    return convertHexToBase64(this.tree.getHexRoot().replace('0x', ''))
  }

  public getMerkleProof(tx: WithdrawalTx): string[] {
    const bridge_id_buf = Buffer.alloc(8)
    bridge_id_buf.writeBigInt64BE(tx.bridge_id)

    const sequence_buf = Buffer.alloc(8)
    sequence_buf.writeBigInt64BE(tx.sequence)

    const amount_buf = Buffer.alloc(8)
    amount_buf.writeBigInt64BE(tx.amount)

    return this.tree
      .getHexProof(
        sha3_256(
          Buffer.concat([
            bridge_id_buf,
            sequence_buf,
            AccAddress.toBuffer(tx.sender),
            AccAddress.toBuffer(tx.receiver),
            Buffer.from(tx.l1_denom, 'utf8'),
            amount_buf
          ])
        )
      )
      .map((v) => convertHexToBase64(v.replace('0x', '')))
  }

  public verify(
    proof: string[],
    tx: {
      bridge_id: bigint;
      sequence: bigint;
      sender: string;
      receiver: string;
      l1_denom: string;
      amount: bigint;
    }
  ): boolean {
    const bridge_id_buf = Buffer.alloc(8)
    bridge_id_buf.writeBigInt64BE(tx.bridge_id)

    const sequence_buf = Buffer.alloc(8)
    sequence_buf.writeBigInt64BE(tx.sequence)

    const amount_buf = Buffer.alloc(8)
    amount_buf.writeBigInt64BE(tx.amount)

    let hashBuf = sha3_256(
      Buffer.concat([
        bridge_id_buf,
        sequence_buf,
        AccAddress.toBuffer(tx.sender),
        AccAddress.toBuffer(tx.receiver),
        Buffer.from(tx.l1_denom, 'utf8'),
        amount_buf
      ])
    )

    proof.forEach((proofElem) => {
      const proofBuf = Buffer.from(proofElem, 'base64')

      if (Buffer.compare(hashBuf, proofBuf) === -1) {
        hashBuf = sha3_256(Buffer.concat([hashBuf, proofBuf]))
      } else {
        hashBuf = sha3_256(Buffer.concat([proofBuf, hashBuf]))
      }
    })

    return this.getMerkleRoot() === hashBuf.toString('base64')
  }
}
