// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/daemon/api/controller/controller.proto (package bifrost.api.controller, syntax proto3)
/* eslint-disable */

import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import { createMessageType, Message } from '@aptre/protobuf-es-lite'
import { Config as Config$1 } from '../api.pb.js'
import { Config as Config$2 } from '@go/github.com/aperturerobotics/controllerbus/bus/api/api.pb.js'

export const protobufPackage = 'bifrost.api.controller'

/**
 * Config configures the API.
 *
 * @generated from message bifrost.api.controller.Config
 */
export type Config = Message<{
  /**
   * ListenAddr is the address to listen on for connections.
   *
   * @generated from field: string listen_addr = 1;
   */
  listenAddr?: string
  /**
   * ApiConfig are api config options.
   *
   * @generated from field: bifrost.api.Config api_config = 2;
   */
  apiConfig?: Config$1
  /**
   * DisableBusApi disables the bus api.
   *
   * @generated from field: bool disable_bus_api = 3;
   */
  disableBusApi?: boolean
  /**
   * BusApiConfig are controller-bus bus api config options.
   * BusApiConfig are options for controller bus api.
   *
   * @generated from field: bus.api.Config bus_api_config = 4;
   */
  busApiConfig?: Config$2
}>

export const Config: MessageType<Config> = createMessageType({
  typeName: 'bifrost.api.controller.Config',
  fields: [
    {
      no: 1,
      name: 'listen_addr',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    { no: 2, name: 'api_config', kind: 'message', T: () => Config$1 },
    {
      no: 3,
      name: 'disable_bus_api',
      kind: 'scalar',
      T: 8 /* ScalarType.BOOL */,
    },
    { no: 4, name: 'bus_api_config', kind: 'message', T: () => Config$2 },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
