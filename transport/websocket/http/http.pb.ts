// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/transport/websocket/http/http.proto (package websocket.http, syntax proto3)
/* eslint-disable */

import { Opts } from '../../common/quic/quic.pb.js'
import { DialerOpts } from '../../common/dialer/dialer.pb.js'
import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import { createMessageType, ScalarType } from '@aptre/protobuf-es-lite'

export const protobufPackage = 'websocket.http'

/**
 * Config is the configuration for the Websocket HTTP handler transport.
 *
 * Listen for incoming connections with bifrost/http/listener
 * This controller resolves LookupHTTPHandler directives filtering by ServeMux patterns.
 * Example: ["GET example.com/my/ws", "GET /other/path"]
 *
 * @generated from message websocket.http.Config
 */
export interface Config {
  /**
   * TransportPeerID sets the peer ID to attach the transport to.
   * If unset, attaches to any running peer with a private key.
   *
   * @generated from field: string transport_peer_id = 1;
   */
  transportPeerId?: string
  /**
   * HttpPatterns is the list of patterns to listen on.
   * Example: ["GET example.com/my/ws", "GET /other/path"]
   *
   * @generated from field: repeated string http_patterns = 2;
   */
  httpPatterns?: string[]
  /**
   * PeerHttpPatterns is the list of patterns to serve the peer ID on.
   * Example: ["GET example.com/my/ws/peer-id", "GET /other/path/peer-id"]
   *
   * @generated from field: repeated string peer_http_patterns = 3;
   */
  peerHttpPatterns?: string[]
  /**
   * Quic contains the quic protocol options.
   *
   * The WebSocket transport always disables FEC and several other UDP-centric
   * features which are unnecessary due to the "reliable" nature of WebSockets.
   *
   * @generated from field: transport.quic.Opts quic = 4;
   */
  quic?: Opts
  /**
   * Dialers maps peer IDs to dialers.
   *
   * @generated from field: map<string, dialer.DialerOpts> dialers = 5;
   */
  dialers?: { [key: string]: DialerOpts }
}

// Config contains the message type declaration for Config.
export const Config: MessageType<Config> = createMessageType({
  typeName: 'websocket.http.Config',
  fields: [
    { no: 1, name: 'transport_peer_id', kind: 'scalar', T: ScalarType.STRING },
    {
      no: 2,
      name: 'http_patterns',
      kind: 'scalar',
      T: ScalarType.STRING,
      repeated: true,
    },
    {
      no: 3,
      name: 'peer_http_patterns',
      kind: 'scalar',
      T: ScalarType.STRING,
      repeated: true,
    },
    { no: 4, name: 'quic', kind: 'message', T: () => Opts },
    {
      no: 5,
      name: 'dialers',
      kind: 'map',
      K: ScalarType.STRING,
      V: { kind: 'message', T: () => DialerOpts },
    },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})