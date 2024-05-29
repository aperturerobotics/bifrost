// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/stream/listening/listening.proto (package stream.listening, syntax proto3)
/* eslint-disable */

import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import { createMessageType, ScalarType } from '@aptre/protobuf-es-lite'

export const protobufPackage = 'stream.listening'

/**
 * Config configures the listening controller.
 *
 * @generated from message stream.listening.Config
 */
export interface Config {
  /**
   * LocalPeerId is the peer ID to forward incoming connections with.
   * Can be empty.
   *
   * @generated from field: string local_peer_id = 1;
   */
  localPeerId?: string
  /**
   * RemotePeerId is the peer ID to forward incoming connections to.
   *
   * @generated from field: string remote_peer_id = 2;
   */
  remotePeerId?: string
  /**
   * ProtocolId is the protocol ID to assign to incoming connections.
   *
   * @generated from field: string protocol_id = 3;
   */
  protocolId?: string
  /**
   * ListenMultiaddr is the listening multiaddress.
   *
   * @generated from field: string listen_multiaddr = 4;
   */
  listenMultiaddr?: string
  /**
   * TransportId sets a transport ID constraint.
   * Can be empty.
   *
   * @generated from field: uint64 transport_id = 5;
   */
  transportId?: bigint
  /**
   * Reliable indicates the stream should be reliable.
   *
   * @generated from field: bool reliable = 6;
   */
  reliable?: boolean
  /**
   * Encrypted indicates the stream should be encrypted.
   *
   * @generated from field: bool encrypted = 7;
   */
  encrypted?: boolean
}

// Config contains the message type declaration for Config.
export const Config: MessageType<Config> = createMessageType({
  typeName: 'stream.listening.Config',
  fields: [
    { no: 1, name: 'local_peer_id', kind: 'scalar', T: ScalarType.STRING },
    { no: 2, name: 'remote_peer_id', kind: 'scalar', T: ScalarType.STRING },
    { no: 3, name: 'protocol_id', kind: 'scalar', T: ScalarType.STRING },
    { no: 4, name: 'listen_multiaddr', kind: 'scalar', T: ScalarType.STRING },
    { no: 5, name: 'transport_id', kind: 'scalar', T: ScalarType.UINT64 },
    { no: 6, name: 'reliable', kind: 'scalar', T: ScalarType.BOOL },
    { no: 7, name: 'encrypted', kind: 'scalar', T: ScalarType.BOOL },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
