
import { DecoratedFastify } from '@mia-platform/custom-plugin-lib'

import { Metrics } from '../config/metrics'
import { EnvironmentVariables } from '../config/env'

export type BaseService = DecoratedFastify<EnvironmentVariables> & {
    // properties always available to services created via Mia-Platform custom-plugin-lib
    customMetrics: Metrics
    config: EnvironmentVariables
}

export type MainService = BaseService & {
    // add here further properties decorated on the service
}
