// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/tptaddr/static/static.proto (package tptaddr.static, syntax proto3)
/* eslint-disable */

import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import { createMessageType, ScalarType } from '@aptre/protobuf-es-lite'

export const protobufPackage = 'tptaddr.static'

/**
 * Config configures the static controller.
 *
 * Handles LookupTptAddr directives with a static list of addresses.
 *
 * @generated from message tptaddr.static.Config
 */
export interface Config {
  /**
   * Addresses is the mapping of peer id to address list.
   *
   * Format: {peer-id}|{transport-id}|{address}
   * Anything after the second | is treated as part of the address.
   *
   * @generated from field: repeated string addresses = 1;
   */
  addresses?: string[]
}

// Config contains the message type declaration for Config.
export const Config: MessageType<Config> = createMessageType({
  typeName: 'tptaddr.static.Config',
  fields: [
    {
      no: 1,
      name: 'addresses',
      kind: 'scalar',
      T: ScalarType.STRING,
      repeated: true,
    },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
