// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/stream/srpc/client/controller/config.proto (package stream.srpc.client.controller, syntax proto3)
/* eslint-disable */

import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import { createMessageType, Message } from '@aptre/protobuf-es-lite'
import { Config as Config$1 } from '../client.pb.js'

export const protobufPackage = 'stream.srpc.client.controller'

/**
 * Config configures mounting a bifrost srpc RPC client to a bus.
 * Resolves the LookupRpcClient directive.
 *
 * @generated from message stream.srpc.client.controller.Config
 */
export type Config = Message<{
  /**
   * Client contains srpc.client configuration for the RPC client.
   *
   * @generated from field: stream.srpc.client.Config client = 1;
   */
  client?: Config$1
  /**
   * ProtocolId is the protocol ID to use to contact the remote RPC service.
   * Must be set.
   *
   * @generated from field: string protocol_id = 2;
   */
  protocolId?: string
  /**
   * ServiceIdPrefixes are the service ID prefixes to match.
   * The prefix will be stripped from the service id before being passed to the client.
   * This is used like: LookupRpcClient<remote/my/service> -> my/service.
   *
   * If empty slice or empty string: matches all LookupRpcClient calls ignoring service ID.
   * Optional.
   *
   * @generated from field: repeated string service_id_prefixes = 3;
   */
  serviceIdPrefixes?: string[]
}>

export const Config: MessageType<Config> = createMessageType({
  typeName: 'stream.srpc.client.controller.Config',
  fields: [
    { no: 1, name: 'client', kind: 'message', T: () => Config$1 },
    {
      no: 2,
      name: 'protocol_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    {
      no: 3,
      name: 'service_id_prefixes',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
      repeated: true,
    },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
