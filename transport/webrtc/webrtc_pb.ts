// @generated by protoc-gen-es v1.8.0 with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/transport/webrtc/webrtc.proto (package webrtc, syntax proto3)
/* eslint-disable */

import type {
  BinaryReadOptions,
  FieldList,
  JsonReadOptions,
  JsonValue,
  PartialMessage,
  PlainMessage,
} from '@bufbuild/protobuf'
import { Message, proto3, protoInt64 } from '@bufbuild/protobuf'
import { Opts } from '../common/quic/quic_pb.js'
import { Backoff } from '../../../util/backoff/backoff_pb.js'
import { DialerOpts } from '../common/dialer/dialer_pb.js'

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
// Retrieve enum metadata with: proto3.getEnumType(IceTransportPolicy)
proto3.util.setEnumType(IceTransportPolicy, 'webrtc.IceTransportPolicy', [
  { no: 0, name: 'IceTransportPolicy_ALL' },
  { no: 1, name: 'IceTransportPolicy_RELAY' },
])

/**
 * Config is the configuration for the WebRTC Signal RPC transport.
 *
 * @generated from message webrtc.Config
 */
export class Config extends Message<Config> {
  /**
   * SignalingId is the signaling channel identifier.
   * Cannot be empty.
   *
   * @generated from field: string signaling_id = 1;
   */
  signalingId = ''

  /**
   * TransportPeerId sets the peer ID to attach the transport to.
   * If unset, attaches to any running peer with a private key.
   * Must match the transport peer ID of the signaling transport.
   *
   * @generated from field: string transport_peer_id = 2;
   */
  transportPeerId = ''

  /**
   * TransportType overrides the transport type id for dial addresses.
   * Defaults to "webrtc"
   * Configures the scheme for addr matching to this transport.
   * E.x.: webrtc://
   *
   * @generated from field: string transport_type = 3;
   */
  transportType = ''

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
  dialers: { [key: string]: DialerOpts } = {}

  /**
   * AllPeers tells the transport to attempt to negotiate a WebRTC session with
   * any peer, even those not listed in the Dialers map.
   *
   * @generated from field: bool all_peers = 8;
   */
  allPeers = false

  /**
   * DisableListen disables listening for incoming Links.
   * If set, we will only dial out, not accept incoming links.
   *
   * @generated from field: bool disable_listen = 9;
   */
  disableListen = false

  /**
   * BlockPeers is a list of peer ids that will not be contacted via this transport.
   *
   * @generated from field: repeated string block_peers = 10;
   */
  blockPeers: string[] = []

  /**
   * Verbose enables very verbose logging.
   *
   * @generated from field: bool verbose = 11;
   */
  verbose = false

  constructor(data?: PartialMessage<Config>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'webrtc.Config'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'signaling_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    {
      no: 2,
      name: 'transport_peer_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    {
      no: 3,
      name: 'transport_type',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    { no: 4, name: 'quic', kind: 'message', T: Opts },
    { no: 5, name: 'web_rtc', kind: 'message', T: WebRtcConfig },
    { no: 6, name: 'backoff', kind: 'message', T: Backoff },
    {
      no: 7,
      name: 'dialers',
      kind: 'map',
      K: 9 /* ScalarType.STRING */,
      V: { kind: 'message', T: DialerOpts },
    },
    { no: 8, name: 'all_peers', kind: 'scalar', T: 8 /* ScalarType.BOOL */ },
    {
      no: 9,
      name: 'disable_listen',
      kind: 'scalar',
      T: 8 /* ScalarType.BOOL */,
    },
    {
      no: 10,
      name: 'block_peers',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
      repeated: true,
    },
    { no: 11, name: 'verbose', kind: 'scalar', T: 8 /* ScalarType.BOOL */ },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): Config {
    return new Config().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): Config {
    return new Config().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): Config {
    return new Config().fromJsonString(jsonString, options)
  }

  static equals(
    a: Config | PlainMessage<Config> | undefined,
    b: Config | PlainMessage<Config> | undefined,
  ): boolean {
    return proto3.util.equals(Config, a, b)
  }
}

/**
 * WebRtcConfig configures the WebRTC PeerConnection.
 *
 * @generated from message webrtc.WebRtcConfig
 */
export class WebRtcConfig extends Message<WebRtcConfig> {
  /**
   * IceServers contains the list of ICE servers to use.
   *
   * @generated from field: repeated webrtc.IceServerConfig ice_servers = 1;
   */
  iceServers: IceServerConfig[] = []

  /**
   * IceTransportPolicy defines the policy for permitted ICE candidates.
   * Optional.
   *
   * @generated from field: webrtc.IceTransportPolicy ice_transport_policy = 2;
   */
  iceTransportPolicy = IceTransportPolicy.IceTransportPolicy_ALL

  /**
   * IceCandidatePoolSize defines the size of the prefetched ICE pool.
   * Optional.
   *
   * @generated from field: uint32 ice_candidate_pool_size = 3;
   */
  iceCandidatePoolSize = 0

  constructor(data?: PartialMessage<WebRtcConfig>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'webrtc.WebRtcConfig'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'ice_servers',
      kind: 'message',
      T: IceServerConfig,
      repeated: true,
    },
    {
      no: 2,
      name: 'ice_transport_policy',
      kind: 'enum',
      T: proto3.getEnumType(IceTransportPolicy),
    },
    {
      no: 3,
      name: 'ice_candidate_pool_size',
      kind: 'scalar',
      T: 13 /* ScalarType.UINT32 */,
    },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): WebRtcConfig {
    return new WebRtcConfig().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): WebRtcConfig {
    return new WebRtcConfig().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): WebRtcConfig {
    return new WebRtcConfig().fromJsonString(jsonString, options)
  }

  static equals(
    a: WebRtcConfig | PlainMessage<WebRtcConfig> | undefined,
    b: WebRtcConfig | PlainMessage<WebRtcConfig> | undefined,
  ): boolean {
    return proto3.util.equals(WebRtcConfig, a, b)
  }
}

/**
 * IceServer is a WebRTC ICE server config.
 *
 * @generated from message webrtc.IceServerConfig
 */
export class IceServerConfig extends Message<IceServerConfig> {
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
  urls: string[] = []

  /**
   * Username is the username for the ICE server.
   *
   * @generated from field: string username = 2;
   */
  username = ''

  /**
   * Credential contains the ice server credential, if any.
   *
   * @generated from oneof webrtc.IceServerConfig.credential
   */
  credential:
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
    | { case: undefined; value?: undefined } = { case: undefined }

  constructor(data?: PartialMessage<IceServerConfig>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'webrtc.IceServerConfig'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'urls',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
      repeated: true,
    },
    { no: 2, name: 'username', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
    {
      no: 3,
      name: 'password',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
      oneof: 'credential',
    },
    {
      no: 4,
      name: 'oauth',
      kind: 'message',
      T: IceServerConfig_OauthCredential,
      oneof: 'credential',
    },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): IceServerConfig {
    return new IceServerConfig().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): IceServerConfig {
    return new IceServerConfig().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): IceServerConfig {
    return new IceServerConfig().fromJsonString(jsonString, options)
  }

  static equals(
    a: IceServerConfig | PlainMessage<IceServerConfig> | undefined,
    b: IceServerConfig | PlainMessage<IceServerConfig> | undefined,
  ): boolean {
    return proto3.util.equals(IceServerConfig, a, b)
  }
}

/**
 * OauthCredential is an OAuth credential information for the ICE server.
 *
 * @generated from message webrtc.IceServerConfig.OauthCredential
 */
export class IceServerConfig_OauthCredential extends Message<IceServerConfig_OauthCredential> {
  /**
   * MacKey is a base64-url format.
   *
   * @generated from field: string mac_key = 1;
   */
  macKey = ''

  /**
   * AccessToken is the access token in base64-encoded format.
   *
   * @generated from field: string access_token = 2;
   */
  accessToken = ''

  constructor(data?: PartialMessage<IceServerConfig_OauthCredential>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'webrtc.IceServerConfig.OauthCredential'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'mac_key', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
    {
      no: 2,
      name: 'access_token',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): IceServerConfig_OauthCredential {
    return new IceServerConfig_OauthCredential().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): IceServerConfig_OauthCredential {
    return new IceServerConfig_OauthCredential().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): IceServerConfig_OauthCredential {
    return new IceServerConfig_OauthCredential().fromJsonString(
      jsonString,
      options,
    )
  }

  static equals(
    a:
      | IceServerConfig_OauthCredential
      | PlainMessage<IceServerConfig_OauthCredential>
      | undefined,
    b:
      | IceServerConfig_OauthCredential
      | PlainMessage<IceServerConfig_OauthCredential>
      | undefined,
  ): boolean {
    return proto3.util.equals(IceServerConfig_OauthCredential, a, b)
  }
}

/**
 * WebRtcSignal is a WebRTC Signaling message sent via the Signaling channel.
 *
 * @generated from message webrtc.WebRtcSignal
 */
export class WebRtcSignal extends Message<WebRtcSignal> {
  /**
   * Body is the body of the message.
   *
   * @generated from oneof webrtc.WebRtcSignal.body
   */
  body:
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
    | { case: undefined; value?: undefined } = { case: undefined }

  constructor(data?: PartialMessage<WebRtcSignal>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'webrtc.WebRtcSignal'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    {
      no: 1,
      name: 'request_offer',
      kind: 'scalar',
      T: 4 /* ScalarType.UINT64 */,
      oneof: 'body',
    },
    { no: 2, name: 'sdp', kind: 'message', T: WebRtcSdp, oneof: 'body' },
    { no: 3, name: 'ice', kind: 'message', T: WebRtcIce, oneof: 'body' },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): WebRtcSignal {
    return new WebRtcSignal().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): WebRtcSignal {
    return new WebRtcSignal().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): WebRtcSignal {
    return new WebRtcSignal().fromJsonString(jsonString, options)
  }

  static equals(
    a: WebRtcSignal | PlainMessage<WebRtcSignal> | undefined,
    b: WebRtcSignal | PlainMessage<WebRtcSignal> | undefined,
  ): boolean {
    return proto3.util.equals(WebRtcSignal, a, b)
  }
}

/**
 * WebRtcSdp contains the SDP offer or answer.
 *
 * @generated from message webrtc.WebRtcSdp
 */
export class WebRtcSdp extends Message<WebRtcSdp> {
  /**
   * TxSeqno is the sequence number of the transmitting peer.
   * The receiver should update the local seqno to match.
   *
   * @generated from field: uint64 tx_seqno = 1;
   */
  txSeqno = protoInt64.zero

  /**
   * SdpType is the string encoded type of the sdp.
   * Examples: "offer" "answer"
   *
   * @generated from field: string sdp_type = 2;
   */
  sdpType = ''

  /**
   * Sdp contains the WebRTC session description.
   *
   * @generated from field: string sdp = 3;
   */
  sdp = ''

  constructor(data?: PartialMessage<WebRtcSdp>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'webrtc.WebRtcSdp'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'tx_seqno', kind: 'scalar', T: 4 /* ScalarType.UINT64 */ },
    { no: 2, name: 'sdp_type', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
    { no: 3, name: 'sdp', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): WebRtcSdp {
    return new WebRtcSdp().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): WebRtcSdp {
    return new WebRtcSdp().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): WebRtcSdp {
    return new WebRtcSdp().fromJsonString(jsonString, options)
  }

  static equals(
    a: WebRtcSdp | PlainMessage<WebRtcSdp> | undefined,
    b: WebRtcSdp | PlainMessage<WebRtcSdp> | undefined,
  ): boolean {
    return proto3.util.equals(WebRtcSdp, a, b)
  }
}

/**
 * WebRtcIce contains an ICE candidate.
 *
 * @generated from message webrtc.WebRtcIce
 */
export class WebRtcIce extends Message<WebRtcIce> {
  /**
   * Candidate contains the JSON-encoded ICE candidate.
   *
   * @generated from field: string candidate = 1;
   */
  candidate = ''

  constructor(data?: PartialMessage<WebRtcIce>) {
    super()
    proto3.util.initPartial(data, this)
  }

  static readonly runtime: typeof proto3 = proto3
  static readonly typeName = 'webrtc.WebRtcIce'
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: 'candidate', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
  ])

  static fromBinary(
    bytes: Uint8Array,
    options?: Partial<BinaryReadOptions>,
  ): WebRtcIce {
    return new WebRtcIce().fromBinary(bytes, options)
  }

  static fromJson(
    jsonValue: JsonValue,
    options?: Partial<JsonReadOptions>,
  ): WebRtcIce {
    return new WebRtcIce().fromJson(jsonValue, options)
  }

  static fromJsonString(
    jsonString: string,
    options?: Partial<JsonReadOptions>,
  ): WebRtcIce {
    return new WebRtcIce().fromJsonString(jsonString, options)
  }

  static equals(
    a: WebRtcIce | PlainMessage<WebRtcIce> | undefined,
    b: WebRtcIce | PlainMessage<WebRtcIce> | undefined,
  ): boolean {
    return proto3.util.equals(WebRtcIce, a, b)
  }
}
