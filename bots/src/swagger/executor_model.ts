import { z } from 'koa-swagger-decorator'

const DepositStruct = z.object({
  bridge_id: z.number(),
  sequence: z.number(),
  l1Denom: z.string(),
  l2Denom: z.string(),
  sender: z.string(),
  receiver: z.string(),
  amount: z.number(),
  outputIndex: z.number(),
  data: z.string(),
  l1Height: z.number()
})

const GetDepositResponse = z.object({
  depositTxList: DepositStruct.array()
})

const WithdrawalStruct = z.object({
  bridge_id: z.number(),
  sequence: z.number(),
  l1Denom: z.string(),
  l2Denom: z.string(),
  sender: z.string(),
  receiver: z.string(),
  amount: z.number(),
  outputIndex: z.number(),
  data: z.string(),
  l1Height: z.number()
})

const GetWithdrawalResponse = z.object({
  withdrawalTxList: WithdrawalStruct.array()
})

const ClaimStruct = z.object({
  bridge_id: z.number(),
  output_index: z.number(),
  merkle_proof: z.string().array(),
  sender: z.string(),
  receiver: z.string(),
  amount: z.number(),
  l_2_denom: z.string(),
  version: z.string(),
  state_root: z.string(),
  merkle_root: z.string(),
  last_block_hash: z.string()
})

const GetClaimResponse = z.object({
  claimTxList: ClaimStruct.array()
})

const OutputStruct = z.object({
  output_index: z.number(),
  output_root: z.string(),
  state_root: z.string(),
  merkle_root: z.string(),
  last_block_hash: z.string(),
  start_block_number: z.number(),
  end_block_number: z.number()
})

const GetOutputResponse = z.object({
  outputList: OutputStruct.array()
})

export {
  GetDepositResponse,
  GetWithdrawalResponse,
  GetClaimResponse,
  GetOutputResponse
}
