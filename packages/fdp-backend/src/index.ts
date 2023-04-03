
import {
  default as customPluginLib,
  CustomService,
  DecoratedFastify,
  ServiceConfig,
} from '@mia-platform/custom-plugin-lib'

import { getMetrics } from './config/metrics'
import { EnvironmentVariables, environmentSchema } from './config/env'
import { BaseService } from './interfaces/service'

const customService: CustomService<EnvironmentVariables> = customPluginLib(environmentSchema)

module.exports = customService(async (service: DecoratedFastify<ServiceConfig>) => {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const baseService = service as BaseService

  /*
   * Insert your code here.
   */
})

// export your custom defined metrics
module.exports.getMetrics = getMetrics
