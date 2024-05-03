// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/stream/api/rpc/rpc.proto (package stream.api.rpc, syntax proto3)
/* eslint-disable */

import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import {
  createEnumType,
  createMessageType,
  Message,
  ScalarType,
} from '@aptre/protobuf-es-lite'

export const protobufPackage = 'stream.api.rpc'

/**
 * StreamState is state for the stream related calls.
 *
 * @generated from enum stream.api.rpc.StreamState
 */
export enum StreamState {
  /**
   * StreamState_NONE indicates nothing about the state
   *
   * @generated from enum value: StreamState_NONE = 0;
   */
  StreamState_NONE = 0,

  /**
   * StreamState_ESTABLISHING indicates the stream is connecting.
   *
   * @generated from enum value: StreamState_ESTABLISHING = 1;
   */
  StreamState_ESTABLISHING = 1,

  /**
   * StreamState_ESTABLISHED indicates the stream is established.
   *
   * @generated from enum value: StreamState_ESTABLISHED = 2;
   */
  StreamState_ESTABLISHED = 2,
}

// StreamState_Enum is the enum type for StreamState.
export const StreamState_Enum = createEnumType('stream.api.rpc.StreamState', [
  { no: 0, name: 'StreamState_NONE' },
  { no: 1, name: 'StreamState_ESTABLISHING' },
  { no: 2, name: 'StreamState_ESTABLISHED' },
])

/**
 * Data is a data packet.
 *
 * @generated from message stream.api.rpc.Data
 */
export type Data = Message<{
  /**
   * State indicates stream state in-band.
   * Data is packet data from the remote.
   *
   * @generated from field: bytes data = 1;
   */
  data?: Uint8Array
  /**
   * State indicates the stream state.
   *
   * @generated from field: stream.api.rpc.StreamState state = 2;
   */
  state?: StreamState
}>

// Data contains the message type declaration for Data.
export const Data: MessageType<Data> = createMessageType({
  typeName: 'stream.api.rpc.Data',
  fields: [
    { no: 1, name: 'data', kind: 'scalar', T: ScalarType.BYTES },
    { no: 2, name: 'state', kind: 'enum', T: StreamState_Enum },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
