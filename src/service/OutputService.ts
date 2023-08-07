import { OutputEntity } from 'orm';
import { getDB } from 'worker/bridgeExecutor/db';
import { APIError, ErrorTypes } from 'lib/error';

export interface GetOutputResponse {
  output: OutputEntity;
}

export interface GetAllOutputsParam {
  offset?: number;
  limit: number;
}

export interface GetAllOutputsResponse {
  next?: number;
  limit: number;
  outputs: OutputEntity[];
}

export interface GetLatestOutputResponse {
  output: OutputEntity;
}

export interface GetOutputByHeightResponse{
  output: OutputEntity;
}


export async function getOutput(
  outputIndex: number
): Promise<GetOutputResponse> {
  const [db] = getDB();
  const queryRunner = db.createQueryRunner('slave');

  try {
    const qb = queryRunner.manager
      .createQueryBuilder(OutputEntity, 'output')
      .where('output.output_index = :outputIndex', { outputIndex });

    const output = await qb.getOne();

    if (!output) {
      throw new APIError(ErrorTypes.NOT_FOUND_ERROR);
    }

    return {
      output
    };
  } finally {
    queryRunner.release();
  }
}

export async function getOutputByHeight(
  height: number
): Promise<GetOutputByHeightResponse> {
  const [db] = getDB();
  const queryRunner = db.createQueryRunner('slave');

  try {
    const qb = queryRunner.manager
      .createQueryBuilder(OutputEntity, 'output')
      .where('output.checkpoint_block_height = :height', { height });

    const output = await qb.getOne();

    if (!output) {
      throw new APIError(ErrorTypes.NOT_FOUND_ERROR);
    }

    return {
      output
    };
  } finally {
    queryRunner.release();
  }
}

export async function getAllOutputs(
  param: GetAllOutputsParam
): Promise<GetAllOutputsResponse> {
  const [db] = getDB();
  const queryRunner = db.createQueryRunner('slave');

  try {
    const offset = param.offset ?? 0;
    const qb = queryRunner.manager
      .createQueryBuilder(OutputEntity, 'output')
      .orderBy('output.outputIndex', 'DESC')
      .skip(offset * param.limit)
      .take(param.limit);

    const outputs = await qb.getMany();
    if (!outputs) {
      throw new APIError(ErrorTypes.NOT_FOUND_ERROR);
    }

    let next: number | undefined;

    // we have next result
    if (param.limit === outputs.length) {
      next = offset + 1;
    }

    return {
      next,
      limit: param.limit,
      outputs
    };
  } finally {
    queryRunner.release();
  }
}

export async function getLatestOutput(): Promise<GetLatestOutputResponse> {
  const [db] = getDB();
  const queryRunner = db.createQueryRunner('slave');

  try {
    const qb = queryRunner.manager
      .createQueryBuilder(OutputEntity, 'output')
      .orderBy('output.outputIndex', 'DESC');

    const output = await qb.getOne();

    if (!output) {
      throw new APIError(ErrorTypes.NOT_FOUND_ERROR);
    }

    return {
      output
    };
  } finally {
    queryRunner.release();
  }
}
