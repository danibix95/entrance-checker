
import { FromSchema } from 'json-schema-to-ts'

export const environmentSchema = {
  type: 'object',
  additionalProperties: {
    type: 'string',
  },
  required: [

  ],
  properties: {
    // add here your environment variables definition
  },
} as const

// allow additional values to be captured in the environment type
export type EnvironmentVariables =
  FromSchema<typeof environmentSchema>
  & { [key: string]: string | number | undefined }
  & (Record<string, string | number> | undefined)
