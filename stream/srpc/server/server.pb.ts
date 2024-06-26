// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/stream/srpc/server/server.proto (package stream.srpc.server, syntax proto3)
/* eslint-disable */

import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import { createMessageType, ScalarType } from '@aptre/protobuf-es-lite'

export const protobufPackage = 'stream.srpc.server'

/**
 * Config configures the server for the srpc service.
 *
 * @generated from message stream.srpc.server.Config
 */
export interface Config {
  /**
   * PeerIds are the list of peer IDs to listen on.
   * If empty, allows any incoming peer id w/ the protocol id(s).
   *
   * @generated from field: repeated string peer_ids = 1;
   */
  peerIds?: string[]
  /**
   * ProtocolIds is the list of protocol ids to listen on.
   * If empty, no incoming streams will be accepted.
   *
   * @generated from field: repeated string protocol_ids = 2;
   */
  protocolIds?: string[]
  /**
   * DisableEstablishLink disables adding an EstablishLink directive for each incoming peer.
   *
   * @generated from field: bool disable_establish_link = 3;
   */
  disableEstablishLink?: boolean
}

// Config contains the message type declaration for Config.
export const Config: MessageType<Config> = createMessageType({
  typeName: 'stream.srpc.server.Config',
  fields: [
    {
      no: 1,
      name: 'peer_ids',
      kind: 'scalar',
      T: ScalarType.STRING,
      repeated: true,
    },
    {
      no: 2,
      name: 'protocol_ids',
      kind: 'scalar',
      T: ScalarType.STRING,
      repeated: true,
    },
    {
      no: 3,
      name: 'disable_establish_link',
      kind: 'scalar',
      T: ScalarType.BOOL,
    },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
