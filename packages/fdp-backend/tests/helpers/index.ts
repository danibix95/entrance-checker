
import lc39 from '@mia-platform/lc39'
import { BaseService } from '../../src/interfaces/service'

import { LOG_LEVEL } from './env'
import { EnvironmentVariables } from '../../src/config/env'

export const startService = async (
  envVariables?: EnvironmentVariables,
): Promise<BaseService> => {
  const service = await lc39('src/index.ts', {
    logLevel: LOG_LEVEL ?? 'silent',
    envVariables,
  }) as BaseService

  await service.ready()

  return service
}
