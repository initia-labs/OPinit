import Axios, { AxiosInstance } from "axios";
import { OutputEntity } from "orm";

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
        const response = await this.api.get<T>(url);
        return response.data;
    }

    public async getOuptut(outputIndx: number): Promise<OutputEntity>{
        return this.getQuery<OutputEntity>(`/output/${outputIndx}`);
    }

    public async getBlock(blockHeight: number): Promise<any> {
        return this.getQuery(`/v1/blocks/${blockHeight}`);
    }
}