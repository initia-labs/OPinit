import { Wallet, MnemonicKey, BCS, Msg, MsgExecute } from '@initia/initia.js';
import axios from 'axios';

import { MoveBuilder } from '@initia/builder.js';
import { getConfig } from 'config';

const config = getConfig();
export const bcs = BCS.getInstance();
export const executor = new Wallet(
  config.l1lcd,
  new MnemonicKey({ mnemonic: config.EXECUTOR_MNEMONIC })
);
export const challenger = new Wallet(
  config.l1lcd,
  new MnemonicKey({ mnemonic: config.CHALLENGER_MNEMONIC })
);
export const outputSubmitter = new Wallet(
  config.l1lcd,
  new MnemonicKey({ mnemonic: config.OUTPUT_SUBMITTER_MNEMONIC })
);

export async function build(
  dirname: string,
  moduleName: string
): Promise<string> {
  const builder = new MoveBuilder(`${dirname}`, {});
  await builder.build();
  const contract = await builder.get(moduleName);
  return contract.toString('base64');
}

export interface TxResponse {
  metadata: string;
  sequence: number;
  sender: string;
  receiver: string;
  amount: number;
  outputIndex: number;
  merkleRoot: string;
  merkleProof: string[];
}

export interface OutputResponse {
  outputIndex: number;
  outputRoot: string;
  stateRoot: string;
  storageRoot: string;
  lastBlockHash: string;
  checkpointBlockHeight: number;
}

export async function makeFinalizeMsg(
  sender: Wallet,
  txRes: TxResponse,
  outputRes: OutputResponse
): Promise<Msg> {
  const msg = new MsgExecute(
    sender.key.accAddress,
    '0x1',
    'op_bridge',
    'finalize_token_bridge',
    [],
    [
      bcs.serialize('address', executor.key.accAddress),
      bcs.serialize('string', config.L2ID),
      bcs.serialize('object', txRes.metadata), // coin metadata
      bcs.serialize('u64', outputRes.outputIndex), // output index
      bcs.serialize(
        'vector<vector<u8>>',
        txRes.merkleProof.map((proof: string) => Buffer.from(proof, 'hex'))
      ), // withdrawal proofs  (tx table)

      // withdraw tx data  (tx table)
      bcs.serialize('u64', txRes.sequence), // l2_sequence (txEntity sequence)
      bcs.serialize('address', txRes.sender), // sender
      bcs.serialize('address', txRes.receiver), // receiver
      bcs.serialize('u64', txRes.amount), // amount

      // output root proof (output table)
      bcs.serialize(
        'vector<u8>',
        Buffer.from(outputRes.outputIndex.toString(), 'utf8')
      ), //version (==output index)
      bcs.serialize('vector<u8>', Buffer.from(outputRes.stateRoot, 'base64')), // state_root
      bcs.serialize('vector<u8>', Buffer.from(outputRes.storageRoot, 'hex')), // storage root
      bcs.serialize(
        'vector<u8>',
        Buffer.from(outputRes.lastBlockHash, 'base64')
      ) // latests block hash
    ]
  );
  return msg;
}

export async function getTx(
  coin: string,
  sequence: number
): Promise<TxResponse> {
  const url = `${config.EXECUTOR_URI}/tx/${coin}/${sequence}`;

  const res = await axios.get(url);
  return res.data;
}

export async function getOutput(outputIndex: number): Promise<OutputResponse> {
  const url = `${config.EXECUTOR_URI}/output/${outputIndex}`;
  const res = await axios.get(url);
  return res.data;
}

export const checkHealth = async (url: string, timeout = 60_000) => {
  const startTime = Date.now();

  while (Date.now() - startTime < timeout) {
    try {
      const response = await axios.get(url);
      if (response.status === 200) return;
    } catch {
      continue;
    }
    await new Promise((res) => setTimeout(res, 1_000));
  }
};
