// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/entitygraph/config.proto (package bifrost.entitygraph, syntax proto3)
/* eslint-disable */

import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import { createMessageType, Message } from '@aptre/protobuf-es-lite'

export const protobufPackage = 'bifrost.entitygraph'

/**
 * Config is the config object for the entitygraph repoter.
 *
 * @generated from message bifrost.entitygraph.Config
 */
export type Config = Message<{}>

// Config contains the message type declaration for Config.
export const Config: MessageType<Config> = createMessageType({
  typeName: 'bifrost.entitygraph.Config',
  fields: [] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
