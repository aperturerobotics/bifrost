// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/signaling/echo/echo.proto (package signaling.echo, syntax proto3)
/* eslint-disable */

import {
  createMessageType,
  Message,
  MessageType,
  PartialFieldInfo,
} from '@aptre/protobuf-es-lite'

export const protobufPackage = 'signaling.echo'

/**
 * Config configures the echo controller.
 *
 * @generated from message signaling.echo.Config
 */
export interface Config extends Message<Config> {
  /**
   * SignalingId is the incoming signaling ID to handle and echo messages.
   * Cannot be empty.
   *
   * @generated from field: string signaling_id = 1;
   */
  signalingId?: string
}

export const Config: MessageType<Config> = createMessageType({
  typeName: 'signaling.echo.Config',
  fields: [
    {
      no: 1,
      name: 'signaling_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
