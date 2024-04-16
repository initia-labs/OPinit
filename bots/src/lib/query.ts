import { BridgeInfo, OutputInfo, TokenPair } from '@initia/initia.js'
import { config } from '../config'
import {
  DepositTxResponse,
  OutputResponse,
  WithdrawalTxResponse
} from './types'
import axios from 'axios'

/// LCD query

// get the latest output from L1 chain
export async function getLastOutputInfo(
  bridgeId: number
): Promise<OutputInfo | null> {
  const [outputInfos] = await config.l1lcd.ophost.outputInfos(bridgeId, {
    'pagination.limit': '1',
    'pagination.reverse': 'true'
  })
  if (outputInfos.length === 0) return null
  return outputInfos[0]
}

// get the output by index from L1 chain
export async function getOutputInfoByIndex(
  bridgeId: number,
  outputIndex: number
): Promise<OutputInfo> {
  return await config.l1lcd.ophost.outputInfo(bridgeId, outputIndex)
}

export async function getBridgeInfo(bridgeId: number): Promise<BridgeInfo> {
  return await config.l1lcd.ophost.bridgeInfo(bridgeId)
}

export async function getTokenPairByL1Denom(denom: string): Promise<TokenPair> {
  return await config.l1lcd.ophost.tokenPairByL1Denom(config.BRIDGE_ID, denom)
}

/// API query

export async function getWithdrawalTxFromExecutor(
  bridge_id: number,
  sequence: number
): Promise<WithdrawalTxResponse> {
  const url = `${config.EXECUTOR_URI}/tx/withdrawal/${bridge_id}/${sequence}`

  const res = await axios.get(url)
  return res.data
}

export async function getDepositTxFromExecutor(
  bridge_id: number,
  sequence: number
): Promise<DepositTxResponse> {
  const url = `${config.EXECUTOR_URI}/tx/deposit/${bridge_id}/${sequence}`
  const res = await axios.get(url)
  return res.data
}

// fetching the output by index from l2 chain
export async function getOutputFromExecutor(
  outputIndex: number
): Promise<OutputResponse> {
  const url = `${config.EXECUTOR_URI}/output/${outputIndex}`
  const res = await axios.get(url)
  return res.data
}

// fetching the latest output from l2 chain
export async function getLatestOutputFromExecutor(): Promise<OutputResponse> {
  const url = `${config.EXECUTOR_URI}/output/latest`
  const res = await axios.get(url)
  return res.data
}

export const checkHealth = async (url: string, timeout = 60_000) => {
  const startTime = Date.now()

  while (Date.now() - startTime < timeout) {
    try {
      const response = await axios.get(url)
      if (response.status === 200) return
    } catch {
      continue
    }
    await new Promise((res) => setTimeout(res, 1_000))
  }
}
