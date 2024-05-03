// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/transport/controller/controller.proto (package transport.controller, syntax proto3)
/* eslint-disable */

import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import { createMessageType, Message, ScalarType } from '@aptre/protobuf-es-lite'

export const protobufPackage = 'transport.controller'

/**
 * StreamEstablish is the first message sent by the initiator of a stream.
 * Prefixed by a uint32 length.
 * Max size: 100kb
 *
 * @generated from message transport.controller.StreamEstablish
 */
export type StreamEstablish = Message<{
  /**
   * ProtocolID is the protocol identifier string for the stream.
   *
   * @generated from field: string protocol_id = 1;
   */
  protocolId?: string
}>

// StreamEstablish contains the message type declaration for StreamEstablish.
export const StreamEstablish: MessageType<StreamEstablish> = createMessageType({
  typeName: 'transport.controller.StreamEstablish',
  fields: [
    { no: 1, name: 'protocol_id', kind: 'scalar', T: ScalarType.STRING },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
