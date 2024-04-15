import * as winston from 'winston'
import axios, { AxiosRequestConfig } from 'axios'
import Websocket from 'ws'

export class RPCSocket {
  public ws: Websocket
  public wsUrl: string
  public sendedPingAt = 0
  public isAlive = true
  public alivedAt: number
  public updateTimer: NodeJS.Timeout
  public latestHeight?: number
  logger: winston.Logger
  rpcUrl: string
  curRPCUrlIndex: number

  constructor(
    public rpcUrls: string[],
    public interval: number,
    logger: winston.Logger
  ) {
    if (this.rpcUrls.length === 0) {
      throw new Error('RPC URLs list cannot be empty')
    }
    this.curRPCUrlIndex = 0
    this.rpcUrl = this.rpcUrls[this.curRPCUrlIndex]
    this.wsUrl = this.rpcUrl.replace('http', 'ws') + '/websocket'
    this.logger = logger
  }

  public initialize(): void {
    this.connect()
    this.updateTimer = setTimeout(() => this.tick(), this.interval)
  }

  public rotateRPC() {
    this.curRPCUrlIndex = (this.curRPCUrlIndex + 1) % this.rpcUrls.length
    this.rpcUrl = this.rpcUrls[this.curRPCUrlIndex]
    this.wsUrl = this.rpcUrl.replace('http', 'ws') + '/websocket'
    this.logger.info(`Rotate WS RPC to ${this.rpcUrl}`)
  }

  public stop(): void {
    if (this.ws) this.ws.terminate()
  }

  public tick(): void {
    const now = Date.now()
    if (
      this.ws &&
      this.ws.readyState === this.ws.OPEN &&
      now - this.sendedPingAt > 10000
    ) {
      this.ws.ping()
      this.sendedPingAt = now
    }

    this.checkAlive()

    if (this.updateTimer) clearTimeout(this.updateTimer)
    this.updateTimer = setTimeout(() => this.tick(), this.interval)
  }

  protected alive(): void {
    if (!this.isAlive) {
      const downtime = (
        (Date.now() - this.alivedAt - this.interval) /
        60 /
        1000
      ).toFixed(1)
      const msg = `${this.constructor.name} is now alive. (downtime ${downtime} minutes)`
      this.logger.info(msg)
      this.isAlive = true
    }
    this.alivedAt = Date.now()
  }

  private checkAlive(): void {
    // no responsed more than 3 minutes, it is down
    if (this.isAlive && Date.now() - this.alivedAt > 3 * 60 * 1000) {
      const msg = `${this.constructor.name} is no response!`
      this.logger.info(msg)
      this.isAlive = false
    }
  }

  public connect(): void {
    this.disconnect()
    this.ws = new Websocket(this.wsUrl)
    this.ws.on('open', () => this.onConnect())
    this.ws.on('close', (code, reason) =>
      this.onDisconnect(code, reason.toString())
    )
    this.ws.on('error', (error) => this.onError(error))
    this.ws.on('message', async (raw) => await this.onRawData(raw))
    this.ws.on('ping', () => this.ws.pong())
    this.ws.on('pong', () => this.alive())
  }

  public disconnect(): void {
    if (this.ws) this.ws.terminate()
  }

  protected onConnect(): void {
    const request = {
      jsonrpc: '2.0',
      method: 'subscribe',
      id: 0,
      params: {
        query: `tm.event = 'NewBlock'`
      }
    }

    this.ws.send(JSON.stringify(request))
    this.logger.info(
      `${this.constructor.name}: websocket connected to ${this.wsUrl}`
    )
    this.alive()
  }

  protected onDisconnect(code: number, reason: string): void {
    this.rotateRPC()
    this.logger.info(
      `${this.constructor.name}: websocket disconnected (${code}: ${reason})`
    )
    // if disconnected, try connect again
    setTimeout(() => this.connect(), 1000)
  }

  // eslint-disable-next-line
  protected onError(error): void {
    this.logger.info(`${this.constructor.name} websocket: `, error)
  }

  // eslint-disable-next-line
  protected async onRawData(raw): Promise<void> {
    let data

    try {
      data = JSON.parse(raw)
    } catch (error) {
      this.logger.info(`${this.constructor.name}: JSON parse error ${raw}`)
      return
    }

    try {
      if (data['result']?.['data']?.['value']) {
        this.latestHeight = Number.parseInt(
          data['result']?.['data']?.['value']['block']['header']['height']
        )
      }
    } catch (error) {
      this.logger.info(error)
    }

    this.alive()
  }
}

export class RPCClient {
  private curRPCUrlIndex = 0
  private rpcUrl: string

  constructor(
    public rpcUrls: string[],
    public logger: winston.Logger
  ) {
    if (this.rpcUrls.length === 0) {
      throw new Error('RPC URLs list cannot be empty')
    }
    this.curRPCUrlIndex = 0
    this.rpcUrl = this.rpcUrls[this.curRPCUrlIndex]
  }

  public rotateRPC() {
    this.curRPCUrlIndex = (this.curRPCUrlIndex + 1) % this.rpcUrls.length
    this.rpcUrl = this.rpcUrls[this.curRPCUrlIndex]
    this.logger.info(`Rotate RPC to ${this.rpcUrl}`)
  }

  async getRequest(
    path: string,
    params?: Record<string, string>
  ): Promise<any> {
    const options: AxiosRequestConfig = {
      headers: {
        'Content-Type': 'application/json',
        'User-Agent': 'initia-rollup'
      }
    }

    let url = `${this.rpcUrl}${path}`
    params &&
      Object.keys(params).forEach(
        (key) => params[key] === undefined && delete params[key]
      )
    const qs = new URLSearchParams(params as any).toString()
    if (qs.length) {
      url += `?${qs}`
    }

    try {
      const response = await axios.get(url, options)
      if (response.status !== 200) {
        throw new Error(`Invalid status code: ${response.status}`)
      }

      const data = response.data
      if (!data || typeof data.jsonrpc !== 'string') {
        throw new Error('Failed to query RPC')
      }

      return data.result
    } catch (e) {
      throw new Error(`RPC request to ${url} failed by ${e}`)
    }
  }

  async getBlockchain(
    min_height: number,
    max_height: number
  ): Promise<Blockchain | null> {
    const blockchainResult: Blockchain = await this.getRequest(`/blockchain`, {
      minHeight: min_height.toString(),
      maxHeight: max_height.toString()
    })

    if (!blockchainResult) {
      this.logger.info('failed get blockchain from rpc')
      return null
    }

    return blockchainResult
  }

  async getBlockBulk(start: string, end: string): Promise<BlockBulk | null> {
    const blockBulksResult: BlockBulk = await this.getRequest(`/block_bulk`, {
      start,
      end
    })

    if (!blockBulksResult) {
      this.logger.info('failed get block bulks from rpc')
      return null
    }

    return blockBulksResult
  }

  async getRawCommit(end: string): Promise<RawCommit | null> {
    const rawCommitResult: RawCommit = await this.getRequest(`/raw_commit`, {
      height: end
    })

    if (!rawCommitResult) {
      this.logger.info('failed get raw commit from rpc')
      return null
    }

    return rawCommitResult
  }

  async lookupInvalidBlock(): Promise<InvalidBlock | null> {
    const invalidBlockResult: InvalidBlock =
      await this.getRequest(`/invalid_block`)

    if (invalidBlockResult.reason !== '' && invalidBlockResult.height !== '0') {
      return invalidBlockResult
    }

    return null
  }

  async getLatestBlockHeight(): Promise<number> {
    const abciInfo: ABCIInfo = await this.getRequest(`/abci_info`)

    if (abciInfo) {
      return parseInt(abciInfo.last_block_height)
    }

    throw new Error(`failed to get latest block height`)
  }
}

export interface Blockchain {
  last_height: string;
  block_metas: BlockMeta[];
}

export interface BlockMeta {
  block_id: any;
  block_size: string;
  header: any;
  num_txs: string;
}
export interface BlockBulk {
  blocks: string[];
}
export interface RawCommit {
  commit: string;
}

interface InvalidBlock {
  reason: string;
  height: string;
}

interface ABCIInfo {
  data: string;
  version: string;
  last_block_height: string;
  last_block_app_hash: string;
}
