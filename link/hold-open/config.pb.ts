// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/link/hold-open/config.proto (package link.holdopen.controller, syntax proto3)
/* eslint-disable */

import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import { createMessageType } from '@aptre/protobuf-es-lite'

export const protobufPackage = 'link.holdopen.controller'

/**
 * Config is the hold-open controller config.
 *
 * TODO: limit to specific transport ID, etc.
 *
 * @generated from message link.holdopen.controller.Config
 */
export interface Config {}

// Config contains the message type declaration for Config.
export const Config: MessageType<Config> = createMessageType({
  typeName: 'link.holdopen.controller.Config',
  fields: [] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
