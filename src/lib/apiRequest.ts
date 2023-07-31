import Axios, { AxiosInstance } from "axios";
import { OutputEntity } from "orm";

interface TxsResponse {
    limit: number;
    txs: Tx[];
    next?: number;
}

export interface OutputProposedEvent {
    l2_id: string;
    output_root: number[],
    output_index: number,
    l2_block_number: number,
    l1_timestamp: number,
}

export interface Tx {
    id: string;
    info: string;
    height: string;
    txhash: string;
    data?: string;
    code?: number;
    codespace?: string;
    raw_log?: string;
    logs: {
        log: string;
        events: Event[];
        msg_index: number;
    }[];
    events: Event[];
    gas_wanted: string;
    gas_used: string;
    tx: any;
    timestamp: string;
}

export interface Event {
    type: string;
    attributes: {
        key: string;
        index?: boolean;
        value: string;
    }[];
}

export class APIRequest {
    private api: AxiosInstance;

    constructor(baseURL: string) {
        this.api = Axios.create({
            baseURL,
            timeout: 30000,
        });
    }

    private async getQuery<T>(url: string): Promise<T> {
        try { 
            const response = await this.api.get<T>(url);
            return response.data;
        } catch(error) {
            throw new Error(`Error in APIRequest: ${error}`);
        }
    }

    public async getOuptut(outputIndx: number): Promise<OutputEntity>{
        return this.getQuery(`/outputs/${outputIndx}`);
    }

    public async getBlock(blockHeight: number): Promise<any> {
        return this.getQuery(`/v1/blocks/${blockHeight}`);
    }

    // private async getChallengablePeriodTxs(query: string): Promise<Tx[]> {
    //     let txsRes = await this.getQuery<TxsResponse>(query);
    //     let txs = txsRes.txs;
    //     const challengePeriod = new Date();
    //     challengePeriod.setDate(challengePeriod.getDate() - 7);

    //     while (txsRes.next) {
    //         txsRes = await this.getQuery<TxsResponse>(`${query}&offset=${txsRes.next}`);
    //         if (new Date(txsRes.txs[0].timestamp) < challengePeriod) {
    //             break;
    //         }
    //         txs = txs.concat(txsRes.txs);
    //     }

    //     return txs.filter((tx) => {
    //         const timestamp = new Date(tx.timestamp);
    //         return timestamp > challengePeriod
    //     }).reverse();
    // }

    // // get transactions not finalized yet. Defaultly, suppose challenge period is 7 days 
    // public async getChallengableTxs(startHeight?: number): Promise<Tx[]> {
    //     const challengePeriod = new Date();
    //     challengePeriod.setDate(challengePeriod.getDate() - 7);
    //     const txs = await this.getChallengablePeriodTxs(`/v1/txs?limit=100`);
    //     if (startHeight){
    //         txs.filter((tx) => {
    //             return Number(tx.height) >= startHeight
    //         })
    //     }
    //     return txs;
    // }

    // public extractChallengableEvents(txs: Tx[], bridgeHexAddress: string, l2ID: string): (OutputProposedEvent| null)[]  {
    //     return txs.flatMap(tx => 
    //         tx.logs.flatMap(log => 
    //             log.events.filter(evt => {
    //                return evt.attributes.some(
    //                     attr => attr.key === "type_tag" 
    //                         && attr.value === `${bridgeHexAddress}::op_output::OutputProposedEvent`
    //                 )
    //             })
    //         )
    //     )
    //     .map(evt => {
    //         const dataAttr = evt.attributes.find(attr => attr.key === 'data');
    //         if (!dataAttr) {
    //             return null
    //         }
            
    //         const data = JSON.parse(dataAttr.value);
    //         if (data.l2_id !== l2ID) {
    //             return null 
    //         }
            
    //         return {
    //             l2_id: data.l2_id,
    //             output_root: data.output_root,
    //             output_index: parseInt(data.output_index),
    //             l2_block_number: parseInt(data.l2_block_number),
    //             l1_timestamp: parseInt(data.l1_timestamp),
    //         }
    //     })
    //     .filter(data => data !== null);
    // }
}