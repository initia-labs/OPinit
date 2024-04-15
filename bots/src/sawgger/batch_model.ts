import { z } from 'koa-swagger-decorator'

const L1DataPathsSchema = z.array(
  z.object({
    index: z.number(),
    txHash: z.string()
  })
)

const L1BatchInfoSchema = z.object({
  type: z.literal('l1'),
  dataPaths: L1DataPathsSchema
})

const CelestiaDataPathsSchema = z.array(
  z.object({
    index: z.number(),
    height: z.number(),
    commitment: z.string()
  })
)

const CelestiaBatchInfoSchema = z.object({
  type: z.literal('celestia'),
  dataPaths: CelestiaDataPathsSchema
})

const BatchInfoStruct = z.union([L1BatchInfoSchema, CelestiaBatchInfoSchema])

const GetBatchResponse = z.object({
  bridge_id: z.number(),
  batch_index: z.number(),
  batch_info: BatchInfoStruct
})

export { GetBatchResponse }
