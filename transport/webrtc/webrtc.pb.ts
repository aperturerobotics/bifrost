// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/transport/webrtc/webrtc.proto (package webrtc, syntax proto3)
/* eslint-disable */

import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import {
  createEnumType,
  createMessageType,
  ScalarType,
} from '@aptre/protobuf-es-lite'
import { Opts } from '../common/quic/quic.pb.js'
import { Backoff } from '@go/github.com/aperturerobotics/util/backoff/backoff.pb.js'
import { DialerOpts } from '../common/dialer/dialer.pb.js'

export const protobufPackage = 'webrtc'

/**
 * IceTransportPolicy contains the set of allowed ICE transport policies.
 *
 * @generated from enum webrtc.IceTransportPolicy
 */
export enum IceTransportPolicy {
  /**
   * IceTransportPolicy_ALL allows any kind of ICE candidate.
   *
   * @generated from enum value: IceTransportPolicy_ALL = 0;
   */
  IceTransportPolicy_ALL = 0,

  /**
   * IceTransportPolicy_RELAY allows only media relay candidates (TURN).
   *
   * @generated from enum value: IceTransportPolicy_RELAY = 1;
   */
  IceTransportPolicy_RELAY = 1,
}

// IceTransportPolicy_Enum is the enum type for IceTransportPolicy.
export const IceTransportPolicy_Enum = createEnumType(
  'webrtc.IceTransportPolicy',
  [
    { no: 0, name: 'IceTransportPolicy_ALL' },
    { no: 1, name: 'IceTransportPolicy_RELAY' },
  ],
)

/**
 * OauthCredential is an OAuth credential information for the ICE server.
 *
 * @generated from message webrtc.IceServerConfig.OauthCredential
 */
export interface IceServerConfig_OauthCredential {
  /**
   * MacKey is a base64-url format.
   *
   * @generated from field: string mac_key = 1;
   */
  macKey?: string
  /**
   * AccessToken is the access token in base64-encoded format.
   *
   * @generated from field: string access_token = 2;
   */
  accessToken?: string
}

// IceServerConfig_OauthCredential contains the message type declaration for IceServerConfig_OauthCredential.
export const IceServerConfig_OauthCredential: MessageType<IceServerConfig_OauthCredential> =
  createMessageType({
    typeName: 'webrtc.IceServerConfig.OauthCredential',
    fields: [
      { no: 1, name: 'mac_key', kind: 'scalar', T: ScalarType.STRING },
      { no: 2, name: 'access_token', kind: 'scalar', T: ScalarType.STRING },
    ] as readonly PartialFieldInfo[],
    packedByDefault: true,
  })

/**
 * IceServer is a WebRTC ICE server config.
 *
 * @generated from message webrtc.IceServerConfig
 */
export interface IceServerConfig {
  /**
   * Urls is the list of URLs for the ICE server.
   *
   * Format: stun:{url} or turn:{url} or turns:{url}?transport=tcp
   * Examples:
   * - stun:stun.l.google.com:19302
   * - stun:stun.stunprotocol.org:3478
   * - turns:google.de?transport=tcp
   *
   * @generated from field: repeated string urls = 1;
   */
  urls?: string[]
  /**
   * Username is the username for the ICE server.
   *
   * @generated from field: string username = 2;
   */
  username?: string

  /**
   * Credential contains the ice server credential, if any.
   *
   * @generated from oneof webrtc.IceServerConfig.credential
   */
  credential?:
    | {
        value?: undefined
        case: undefined
      }
    | {
        /**
         * Password contains the ICE server password.
         *
         * @generated from field: string password = 3;
         */
        value: string
        case: 'password'
      }
    | {
        /**
         * Oauth contains an OAuth credential.
         *
         * @generated from field: webrtc.IceServerConfig.OauthCredential oauth = 4;
         */
        value: IceServerConfig_OauthCredential
        case: 'oauth'
      }
}

// IceServerConfig contains the message type declaration for IceServerConfig.
export const IceServerConfig: MessageType<IceServerConfig> = createMessageType({
  typeName: 'webrtc.IceServerConfig',
  fields: [
    {
      no: 1,
      name: 'urls',
      kind: 'scalar',
      T: ScalarType.STRING,
      repeated: true,
    },
    { no: 2, name: 'username', kind: 'scalar', T: ScalarType.STRING },
    {
      no: 3,
      name: 'password',
      kind: 'scalar',
      T: ScalarType.STRING,
      oneof: 'credential',
    },
    {
      no: 4,
      name: 'oauth',
      kind: 'message',
      T: () => IceServerConfig_OauthCredential,
      oneof: 'credential',
    },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})

/**
 * WebRtcConfig configures the WebRTC PeerConnection.
 *
 * @generated from message webrtc.WebRtcConfig
 */
export interface WebRtcConfig {
  /**
   * IceServers contains the list of ICE servers to use.
   *
   * @generated from field: repeated webrtc.IceServerConfig ice_servers = 1;
   */
  iceServers?: IceServerConfig[]
  /**
   * IceTransportPolicy defines the policy for permitted ICE candidates.
   * Optional.
   *
   * @generated from field: webrtc.IceTransportPolicy ice_transport_policy = 2;
   */
  iceTransportPolicy?: IceTransportPolicy
  /**
   * IceCandidatePoolSize defines the size of the prefetched ICE pool.
   * Optional.
   *
   * @generated from field: uint32 ice_candidate_pool_size = 3;
   */
  iceCandidatePoolSize?: number
}

// WebRtcConfig contains the message type declaration for WebRtcConfig.
export const WebRtcConfig: MessageType<WebRtcConfig> = createMessageType({
  typeName: 'webrtc.WebRtcConfig',
  fields: [
    {
      no: 1,
      name: 'ice_servers',
      kind: 'message',
      T: () => IceServerConfig,
      repeated: true,
    },
    {
      no: 2,
      name: 'ice_transport_policy',
      kind: 'enum',
      T: IceTransportPolicy_Enum,
    },
    {
      no: 3,
      name: 'ice_candidate_pool_size',
      kind: 'scalar',
      T: ScalarType.UINT32,
    },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})

/**
 * Config is the configuration for the WebRTC Signal RPC transport.
 *
 * @generated from message webrtc.Config
 */
export interface Config {
  /**
   * SignalingId is the signaling channel identifier.
   * Cannot be empty.
   *
   * @generated from field: string signaling_id = 1;
   */
  signalingId?: string
  /**
   * TransportPeerId sets the peer ID to attach the transport to.
   * If unset, attaches to any running peer with a private key.
   * Must match the transport peer ID of the signaling transport.
   *
   * @generated from field: string transport_peer_id = 2;
   */
  transportPeerId?: string
  /**
   * TransportType overrides the transport type id for dial addresses.
   * Defaults to "webrtc"
   * Configures the scheme for addr matching to this transport.
   * E.x.: webrtc://
   *
   * @generated from field: string transport_type = 3;
   */
  transportType?: string
  /**
   * Quic contains the quic protocol options.
   *
   * The WebRTC transport always disables FEC and several other UDP-centric
   * features which are unnecessary due to the "reliable" nature of WebRTC.
   *
   * @generated from field: transport.quic.Opts quic = 4;
   */
  quic?: Opts
  /**
   * WebRtc contains the WebRTC protocol options.
   *
   * @generated from field: webrtc.WebRtcConfig web_rtc = 5;
   */
  webRtc?: WebRtcConfig
  /**
   * Backoff is the backoff config for connecting to a PeerConnection.
   * If unset, defaults to reasonable defaults.
   *
   * @generated from field: backoff.Backoff backoff = 6;
   */
  backoff?: Backoff
  /**
   * Dialers maps peer IDs to dialers.
   * This allows mapping which peer ID should be dialed via this transport.
   *
   * @generated from field: map<string, dialer.DialerOpts> dialers = 7;
   */
  dialers?: { [key: string]: DialerOpts }
  /**
   * AllPeers tells the transport to attempt to negotiate a WebRTC session with
   * any peer, even those not listed in the Dialers map.
   *
   * @generated from field: bool all_peers = 8;
   */
  allPeers?: boolean
  /**
   * DisableListen disables listening for incoming Links.
   * If set, we will only dial out, not accept incoming links.
   *
   * @generated from field: bool disable_listen = 9;
   */
  disableListen?: boolean
  /**
   * BlockPeers is a list of peer ids that will not be contacted via this transport.
   *
   * @generated from field: repeated string block_peers = 10;
   */
  blockPeers?: string[]
  /**
   * Verbose enables very verbose logging.
   *
   * @generated from field: bool verbose = 11;
   */
  verbose?: boolean
}

// Config contains the message type declaration for Config.
export const Config: MessageType<Config> = createMessageType({
  typeName: 'webrtc.Config',
  fields: [
    { no: 1, name: 'signaling_id', kind: 'scalar', T: ScalarType.STRING },
    { no: 2, name: 'transport_peer_id', kind: 'scalar', T: ScalarType.STRING },
    { no: 3, name: 'transport_type', kind: 'scalar', T: ScalarType.STRING },
    { no: 4, name: 'quic', kind: 'message', T: () => Opts },
    { no: 5, name: 'web_rtc', kind: 'message', T: () => WebRtcConfig },
    { no: 6, name: 'backoff', kind: 'message', T: () => Backoff },
    {
      no: 7,
      name: 'dialers',
      kind: 'map',
      K: ScalarType.STRING,
      V: { kind: 'message', T: () => DialerOpts },
    },
    { no: 8, name: 'all_peers', kind: 'scalar', T: ScalarType.BOOL },
    { no: 9, name: 'disable_listen', kind: 'scalar', T: ScalarType.BOOL },
    {
      no: 10,
      name: 'block_peers',
      kind: 'scalar',
      T: ScalarType.STRING,
      repeated: true,
    },
    { no: 11, name: 'verbose', kind: 'scalar', T: ScalarType.BOOL },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})

/**
 * WebRtcSdp contains the SDP offer or answer.
 *
 * @generated from message webrtc.WebRtcSdp
 */
export interface WebRtcSdp {
  /**
   * TxSeqno is the sequence number of the transmitting peer.
   * The receiver should update the local seqno to match.
   *
   * @generated from field: uint64 tx_seqno = 1;
   */
  txSeqno?: bigint
  /**
   * SdpType is the string encoded type of the sdp.
   * Examples: "offer" "answer"
   *
   * @generated from field: string sdp_type = 2;
   */
  sdpType?: string
  /**
   * Sdp contains the WebRTC session description.
   *
   * @generated from field: string sdp = 3;
   */
  sdp?: string
}

// WebRtcSdp contains the message type declaration for WebRtcSdp.
export const WebRtcSdp: MessageType<WebRtcSdp> = createMessageType({
  typeName: 'webrtc.WebRtcSdp',
  fields: [
    { no: 1, name: 'tx_seqno', kind: 'scalar', T: ScalarType.UINT64 },
    { no: 2, name: 'sdp_type', kind: 'scalar', T: ScalarType.STRING },
    { no: 3, name: 'sdp', kind: 'scalar', T: ScalarType.STRING },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})

/**
 * WebRtcIce contains an ICE candidate.
 *
 * @generated from message webrtc.WebRtcIce
 */
export interface WebRtcIce {
  /**
   * Candidate contains the JSON-encoded ICE candidate.
   *
   * @generated from field: string candidate = 1;
   */
  candidate?: string
}

// WebRtcIce contains the message type declaration for WebRtcIce.
export const WebRtcIce: MessageType<WebRtcIce> = createMessageType({
  typeName: 'webrtc.WebRtcIce',
  fields: [
    { no: 1, name: 'candidate', kind: 'scalar', T: ScalarType.STRING },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})

/**
 * WebRtcSignal is a WebRTC Signaling message sent via the Signaling channel.
 *
 * @generated from message webrtc.WebRtcSignal
 */
export interface WebRtcSignal {
  /**
   * Body is the body of the message.
   *
   * @generated from oneof webrtc.WebRtcSignal.body
   */
  body?:
    | {
        value?: undefined
        case: undefined
      }
    | {
        /**
         * RequestOffer requests a new offer from the offerer with the local session seqno.
         * Incremented when negotiation is needed (something changes about the session).
         *
         * @generated from field: uint64 request_offer = 1;
         */
        value: bigint
        case: 'requestOffer'
      }
    | {
        /**
         * Sdp contains the sdp offer or answer.
         *
         * @generated from field: webrtc.WebRtcSdp sdp = 2;
         */
        value: WebRtcSdp
        case: 'sdp'
      }
    | {
        /**
         * Ice contains an ICE candidate.
         *
         * @generated from field: webrtc.WebRtcIce ice = 3;
         */
        value: WebRtcIce
        case: 'ice'
      }
}

// WebRtcSignal contains the message type declaration for WebRtcSignal.
export const WebRtcSignal: MessageType<WebRtcSignal> = createMessageType({
  typeName: 'webrtc.WebRtcSignal',
  fields: [
    {
      no: 1,
      name: 'request_offer',
      kind: 'scalar',
      T: ScalarType.UINT64,
      oneof: 'body',
    },
    { no: 2, name: 'sdp', kind: 'message', T: () => WebRtcSdp, oneof: 'body' },
    { no: 3, name: 'ice', kind: 'message', T: () => WebRtcIce, oneof: 'body' },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
