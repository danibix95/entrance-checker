
import { EnvironmentVariables } from '../../src/config/env'

export const getConfig = async (): Promise<EnvironmentVariables> => {
  return {
    // Mia - Custom Plugin Lib
    USERID_HEADER_KEY: 'userid',
    GROUPS_HEADER_KEY: 'groups',
    CLIENTTYPE_HEADER_KEY: 'clienttype',
    BACKOFFICE_HEADER_KEY: 'backoffice',
    MICROSERVICE_GATEWAY_SERVICE_NAME: 'microservice-gateway',
    ADDITIONAL_HEADERS_TO_PROXY: 'console-sid,cms-sid',

    // Service Related
  }
}

export const LOG_LEVEL = 'silent'
