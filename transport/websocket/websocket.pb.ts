// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/transport/websocket/websocket.proto (package websocket, syntax proto3)
/* eslint-disable */

import { Opts } from '../common/quic/quic.pb.js'
import { DialerOpts } from '../common/dialer/dialer.pb.js'
import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import { createMessageType, ScalarType } from '@aptre/protobuf-es-lite'

export const protobufPackage = 'websocket'

/**
 * Config is the configuration for the Websocket transport.
 *
 * Quic is used for mTLS mutual authentication over the Websocket, as well as
 * congestion control, stream multiplexing, and other performance features.
 *
 * @generated from message websocket.Config
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
   * ListenAddr contains the address to listen on.
   * Has no effect in the browser.
   *
   * @generated from field: string listen_addr = 2;
   */
  listenAddr?: string
  /**
   * Quic contains the quic protocol options.
   *
   * The WebSocket transport always disables FEC and several other UDP-centric
   * features which are unnecessary due to the "reliable" nature of WebSockets.
   *
   * @generated from field: transport.quic.Opts quic = 3;
   */
  quic?: Opts
  /**
   * Dialers maps peer IDs to dialers.
   *
   * @generated from field: map<string, dialer.DialerOpts> dialers = 4;
   */
  dialers?: { [key: string]: DialerOpts }
  /**
   * HttpPath is the http path to expose the websocket.
   * If unset, disables serving the websocket.
   *
   * @generated from field: string http_path = 5;
   */
  httpPath?: string
  /**
   * HttpPeerPath is the http path to expose the peer id.
   * If unset, disables serving the websocket peer id.
   *
   * @generated from field: string http_peer_path = 6;
   */
  httpPeerPath?: string
  /**
   * Verbose enables verbose logging.
   *
   * @generated from field: bool verbose = 7;
   */
  verbose?: boolean
}

// Config contains the message type declaration for Config.
export const Config: MessageType<Config> = createMessageType({
  typeName: 'websocket.Config',
  fields: [
    { no: 1, name: 'transport_peer_id', kind: 'scalar', T: ScalarType.STRING },
    { no: 2, name: 'listen_addr', kind: 'scalar', T: ScalarType.STRING },
    { no: 3, name: 'quic', kind: 'message', T: () => Opts },
    {
      no: 4,
      name: 'dialers',
      kind: 'map',
      K: ScalarType.STRING,
      V: { kind: 'message', T: () => DialerOpts },
    },
    { no: 5, name: 'http_path', kind: 'scalar', T: ScalarType.STRING },
    { no: 6, name: 'http_peer_path', kind: 'scalar', T: ScalarType.STRING },
    { no: 7, name: 'verbose', kind: 'scalar', T: ScalarType.BOOL },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
