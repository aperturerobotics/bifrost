/* eslint-disable */
import { Backoff } from '@go/github.com/aperturerobotics/util/backoff/backoff.pb.js'
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import { DialerOpts } from '../common/dialer/dialer.pb.js'
import { Opts } from '../common/quic/quic.pb.js'

export const protobufPackage = 'webrtc'

/** IceTransportPolicy contains the set of allowed ICE transport policies. */
export enum IceTransportPolicy {
  /** IceTransportPolicy_ALL - IceTransportPolicy_ALL allows any kind of ICE candidate. */
  IceTransportPolicy_ALL = 0,
  /** IceTransportPolicy_RELAY - IceTransportPolicy_RELAY allows only media relay candidates (TURN). */
  IceTransportPolicy_RELAY = 1,
  UNRECOGNIZED = -1,
}

export function iceTransportPolicyFromJSON(object: any): IceTransportPolicy {
  switch (object) {
    case 0:
    case 'IceTransportPolicy_ALL':
      return IceTransportPolicy.IceTransportPolicy_ALL
    case 1:
    case 'IceTransportPolicy_RELAY':
      return IceTransportPolicy.IceTransportPolicy_RELAY
    case -1:
    case 'UNRECOGNIZED':
    default:
      return IceTransportPolicy.UNRECOGNIZED
  }
}

export function iceTransportPolicyToJSON(object: IceTransportPolicy): string {
  switch (object) {
    case IceTransportPolicy.IceTransportPolicy_ALL:
      return 'IceTransportPolicy_ALL'
    case IceTransportPolicy.IceTransportPolicy_RELAY:
      return 'IceTransportPolicy_RELAY'
    case IceTransportPolicy.UNRECOGNIZED:
    default:
      return 'UNRECOGNIZED'
  }
}

/** Config is the configuration for the WebRTC Signal RPC transport. */
export interface Config {
  /**
   * SignalingId is the signaling channel identifier.
   * Cannot be empty.
   */
  signalingId: string
  /**
   * TransportPeerId sets the peer ID to attach the transport to.
   * If unset, attaches to any running peer with a private key.
   * Must match the transport peer ID of the signaling transport.
   */
  transportPeerId: string
  /**
   * TransportType overrides the transport type id for dial addresses.
   * Defaults to "webrtc"
   * Configures the scheme for addr matching to this transport.
   * E.x.: webrtc://
   */
  transportType: string
  /**
   * Quic contains the quic protocol options.
   *
   * The WebRTC transport always disables FEC and several other UDP-centric
   * features which are unnecessary due to the "reliable" nature of WebRTC.
   */
  quic: Opts | undefined
  /** WebRtc contains the WebRTC protocol options. */
  webRtc: WebRtcConfig | undefined
  /**
   * Backoff is the backoff config for connecting to a PeerConnection.
   * If unset, defaults to reasonable defaults.
   */
  backoff: Backoff | undefined
  /**
   * Dialers maps peer IDs to dialers.
   * This allows mapping which peer ID should be dialed via this transport.
   */
  dialers: { [key: string]: DialerOpts }
  /**
   * AllPeers tells the transport to attempt to negotiate a WebRTC session with
   * any peer, even those not listed in the Dialers map.
   */
  allPeers: boolean
  /**
   * DisableListen disables listening for incoming Links.
   * If set, we will only dial out, not accept incoming links.
   */
  disableListen: boolean
  /** BlockPeers is a list of peer ids that will not be contacted via this transport. */
  blockPeers: string[]
  /** Verbose enables very verbose logging. */
  verbose: boolean
}

export interface Config_DialersEntry {
  key: string
  value: DialerOpts | undefined
}

/** WebRtcConfig configures the WebRTC PeerConnection. */
export interface WebRtcConfig {
  /** IceServers contains the list of ICE servers to use. */
  iceServers: IceServerConfig[]
  /**
   * IceTransportPolicy defines the policy for permitted ICE candidates.
   * Optional.
   */
  iceTransportPolicy: IceTransportPolicy
  /**
   * IceCandidatePoolSize defines the size of the prefetched ICE pool.
   * Optional.
   */
  iceCandidatePoolSize: number
}

/** IceServer is a WebRTC ICE server config. */
export interface IceServerConfig {
  /**
   * Urls is the list of URLs for the ICE server.
   *
   * Format: stun:{url} or turn:{url} or turns:{url}?transport=tcp
   * Examples:
   * - stun:stun.l.google.com:19302
   * - stun:stun.stunprotocol.org:3478
   * - turns:google.de?transport=tcp
   */
  urls: string[]
  /** Username is the username for the ICE server. */
  username: string
  credential?:
    | { $case: 'password'; password: string }
    | { $case: 'oauth'; oauth: IceServerConfig_OauthCredential }
    | undefined
}

/** OauthCredential is an OAuth credential information for the ICE server. */
export interface IceServerConfig_OauthCredential {
  /** MacKey is a base64-url format. */
  macKey: string
  /** AccessToken is the access token in base64-encoded format. */
  accessToken: string
}

/** WebRtcSignal is a WebRTC Signaling message sent via the Signaling channel. */
export interface WebRtcSignal {
  body?:
    | { $case: 'requestOffer'; requestOffer: Long }
    | { $case: 'sdp'; sdp: WebRtcSdp }
    | {
        $case: 'ice'
        ice: WebRtcIce
      }
    | undefined
}

/** WebRtcSdp contains the SDP offer or answer. */
export interface WebRtcSdp {
  /**
   * TxSeqno is the sequence number of the transmitting peer.
   * The receiver should update the local seqno to match.
   */
  txSeqno: Long
  /**
   * SdpType is the string encoded type of the sdp.
   * Examples: "offer" "answer"
   */
  sdpType: string
  /** Sdp contains the WebRTC session description. */
  sdp: string
}

/** WebRtcIce contains an ICE candidate. */
export interface WebRtcIce {
  /** Candidate contains the JSON-encoded ICE candidate. */
  candidate: string
}

function createBaseConfig(): Config {
  return {
    signalingId: '',
    transportPeerId: '',
    transportType: '',
    quic: undefined,
    webRtc: undefined,
    backoff: undefined,
    dialers: {},
    allPeers: false,
    disableListen: false,
    blockPeers: [],
    verbose: false,
  }
}

export const Config = {
  encode(
    message: Config,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.signalingId !== '') {
      writer.uint32(10).string(message.signalingId)
    }
    if (message.transportPeerId !== '') {
      writer.uint32(18).string(message.transportPeerId)
    }
    if (message.transportType !== '') {
      writer.uint32(26).string(message.transportType)
    }
    if (message.quic !== undefined) {
      Opts.encode(message.quic, writer.uint32(34).fork()).ldelim()
    }
    if (message.webRtc !== undefined) {
      WebRtcConfig.encode(message.webRtc, writer.uint32(42).fork()).ldelim()
    }
    if (message.backoff !== undefined) {
      Backoff.encode(message.backoff, writer.uint32(50).fork()).ldelim()
    }
    Object.entries(message.dialers).forEach(([key, value]) => {
      Config_DialersEntry.encode(
        { key: key as any, value },
        writer.uint32(58).fork(),
      ).ldelim()
    })
    if (message.allPeers !== false) {
      writer.uint32(64).bool(message.allPeers)
    }
    if (message.disableListen !== false) {
      writer.uint32(72).bool(message.disableListen)
    }
    for (const v of message.blockPeers) {
      writer.uint32(82).string(v!)
    }
    if (message.verbose !== false) {
      writer.uint32(88).bool(message.verbose)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Config {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseConfig()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.signalingId = reader.string()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.transportPeerId = reader.string()
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.transportType = reader.string()
          continue
        case 4:
          if (tag !== 34) {
            break
          }

          message.quic = Opts.decode(reader, reader.uint32())
          continue
        case 5:
          if (tag !== 42) {
            break
          }

          message.webRtc = WebRtcConfig.decode(reader, reader.uint32())
          continue
        case 6:
          if (tag !== 50) {
            break
          }

          message.backoff = Backoff.decode(reader, reader.uint32())
          continue
        case 7:
          if (tag !== 58) {
            break
          }

          const entry7 = Config_DialersEntry.decode(reader, reader.uint32())
          if (entry7.value !== undefined) {
            message.dialers[entry7.key] = entry7.value
          }
          continue
        case 8:
          if (tag !== 64) {
            break
          }

          message.allPeers = reader.bool()
          continue
        case 9:
          if (tag !== 72) {
            break
          }

          message.disableListen = reader.bool()
          continue
        case 10:
          if (tag !== 82) {
            break
          }

          message.blockPeers.push(reader.string())
          continue
        case 11:
          if (tag !== 88) {
            break
          }

          message.verbose = reader.bool()
          continue
      }
      if ((tag & 7) === 4 || tag === 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Config, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<Config | Config[]> | Iterable<Config | Config[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Config.encode(p).finish()]
        }
      } else {
        yield* [Config.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Config>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<Config> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Config.decode(p)]
        }
      } else {
        yield* [Config.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): Config {
    return {
      signalingId: isSet(object.signalingId)
        ? globalThis.String(object.signalingId)
        : '',
      transportPeerId: isSet(object.transportPeerId)
        ? globalThis.String(object.transportPeerId)
        : '',
      transportType: isSet(object.transportType)
        ? globalThis.String(object.transportType)
        : '',
      quic: isSet(object.quic) ? Opts.fromJSON(object.quic) : undefined,
      webRtc: isSet(object.webRtc)
        ? WebRtcConfig.fromJSON(object.webRtc)
        : undefined,
      backoff: isSet(object.backoff)
        ? Backoff.fromJSON(object.backoff)
        : undefined,
      dialers: isObject(object.dialers)
        ? Object.entries(object.dialers).reduce<{ [key: string]: DialerOpts }>(
            (acc, [key, value]) => {
              acc[key] = DialerOpts.fromJSON(value)
              return acc
            },
            {},
          )
        : {},
      allPeers: isSet(object.allPeers)
        ? globalThis.Boolean(object.allPeers)
        : false,
      disableListen: isSet(object.disableListen)
        ? globalThis.Boolean(object.disableListen)
        : false,
      blockPeers: globalThis.Array.isArray(object?.blockPeers)
        ? object.blockPeers.map((e: any) => globalThis.String(e))
        : [],
      verbose: isSet(object.verbose)
        ? globalThis.Boolean(object.verbose)
        : false,
    }
  },

  toJSON(message: Config): unknown {
    const obj: any = {}
    if (message.signalingId !== '') {
      obj.signalingId = message.signalingId
    }
    if (message.transportPeerId !== '') {
      obj.transportPeerId = message.transportPeerId
    }
    if (message.transportType !== '') {
      obj.transportType = message.transportType
    }
    if (message.quic !== undefined) {
      obj.quic = Opts.toJSON(message.quic)
    }
    if (message.webRtc !== undefined) {
      obj.webRtc = WebRtcConfig.toJSON(message.webRtc)
    }
    if (message.backoff !== undefined) {
      obj.backoff = Backoff.toJSON(message.backoff)
    }
    if (message.dialers) {
      const entries = Object.entries(message.dialers)
      if (entries.length > 0) {
        obj.dialers = {}
        entries.forEach(([k, v]) => {
          obj.dialers[k] = DialerOpts.toJSON(v)
        })
      }
    }
    if (message.allPeers !== false) {
      obj.allPeers = message.allPeers
    }
    if (message.disableListen !== false) {
      obj.disableListen = message.disableListen
    }
    if (message.blockPeers?.length) {
      obj.blockPeers = message.blockPeers
    }
    if (message.verbose !== false) {
      obj.verbose = message.verbose
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Config>, I>>(base?: I): Config {
    return Config.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig()
    message.signalingId = object.signalingId ?? ''
    message.transportPeerId = object.transportPeerId ?? ''
    message.transportType = object.transportType ?? ''
    message.quic =
      object.quic !== undefined && object.quic !== null
        ? Opts.fromPartial(object.quic)
        : undefined
    message.webRtc =
      object.webRtc !== undefined && object.webRtc !== null
        ? WebRtcConfig.fromPartial(object.webRtc)
        : undefined
    message.backoff =
      object.backoff !== undefined && object.backoff !== null
        ? Backoff.fromPartial(object.backoff)
        : undefined
    message.dialers = Object.entries(object.dialers ?? {}).reduce<{
      [key: string]: DialerOpts
    }>((acc, [key, value]) => {
      if (value !== undefined) {
        acc[key] = DialerOpts.fromPartial(value)
      }
      return acc
    }, {})
    message.allPeers = object.allPeers ?? false
    message.disableListen = object.disableListen ?? false
    message.blockPeers = object.blockPeers?.map((e) => e) || []
    message.verbose = object.verbose ?? false
    return message
  },
}

function createBaseConfig_DialersEntry(): Config_DialersEntry {
  return { key: '', value: undefined }
}

export const Config_DialersEntry = {
  encode(
    message: Config_DialersEntry,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.key !== '') {
      writer.uint32(10).string(message.key)
    }
    if (message.value !== undefined) {
      DialerOpts.encode(message.value, writer.uint32(18).fork()).ldelim()
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Config_DialersEntry {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseConfig_DialersEntry()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.key = reader.string()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.value = DialerOpts.decode(reader, reader.uint32())
          continue
      }
      if ((tag & 7) === 4 || tag === 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Config_DialersEntry, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<Config_DialersEntry | Config_DialersEntry[]>
      | Iterable<Config_DialersEntry | Config_DialersEntry[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Config_DialersEntry.encode(p).finish()]
        }
      } else {
        yield* [Config_DialersEntry.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Config_DialersEntry>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<Config_DialersEntry> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [Config_DialersEntry.decode(p)]
        }
      } else {
        yield* [Config_DialersEntry.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): Config_DialersEntry {
    return {
      key: isSet(object.key) ? globalThis.String(object.key) : '',
      value: isSet(object.value)
        ? DialerOpts.fromJSON(object.value)
        : undefined,
    }
  },

  toJSON(message: Config_DialersEntry): unknown {
    const obj: any = {}
    if (message.key !== '') {
      obj.key = message.key
    }
    if (message.value !== undefined) {
      obj.value = DialerOpts.toJSON(message.value)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Config_DialersEntry>, I>>(
    base?: I,
  ): Config_DialersEntry {
    return Config_DialersEntry.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<Config_DialersEntry>, I>>(
    object: I,
  ): Config_DialersEntry {
    const message = createBaseConfig_DialersEntry()
    message.key = object.key ?? ''
    message.value =
      object.value !== undefined && object.value !== null
        ? DialerOpts.fromPartial(object.value)
        : undefined
    return message
  },
}

function createBaseWebRtcConfig(): WebRtcConfig {
  return { iceServers: [], iceTransportPolicy: 0, iceCandidatePoolSize: 0 }
}

export const WebRtcConfig = {
  encode(
    message: WebRtcConfig,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.iceServers) {
      IceServerConfig.encode(v!, writer.uint32(10).fork()).ldelim()
    }
    if (message.iceTransportPolicy !== 0) {
      writer.uint32(16).int32(message.iceTransportPolicy)
    }
    if (message.iceCandidatePoolSize !== 0) {
      writer.uint32(24).uint32(message.iceCandidatePoolSize)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): WebRtcConfig {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseWebRtcConfig()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.iceServers.push(
            IceServerConfig.decode(reader, reader.uint32()),
          )
          continue
        case 2:
          if (tag !== 16) {
            break
          }

          message.iceTransportPolicy = reader.int32() as any
          continue
        case 3:
          if (tag !== 24) {
            break
          }

          message.iceCandidatePoolSize = reader.uint32()
          continue
      }
      if ((tag & 7) === 4 || tag === 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<WebRtcConfig, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<WebRtcConfig | WebRtcConfig[]>
      | Iterable<WebRtcConfig | WebRtcConfig[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [WebRtcConfig.encode(p).finish()]
        }
      } else {
        yield* [WebRtcConfig.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, WebRtcConfig>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<WebRtcConfig> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [WebRtcConfig.decode(p)]
        }
      } else {
        yield* [WebRtcConfig.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): WebRtcConfig {
    return {
      iceServers: globalThis.Array.isArray(object?.iceServers)
        ? object.iceServers.map((e: any) => IceServerConfig.fromJSON(e))
        : [],
      iceTransportPolicy: isSet(object.iceTransportPolicy)
        ? iceTransportPolicyFromJSON(object.iceTransportPolicy)
        : 0,
      iceCandidatePoolSize: isSet(object.iceCandidatePoolSize)
        ? globalThis.Number(object.iceCandidatePoolSize)
        : 0,
    }
  },

  toJSON(message: WebRtcConfig): unknown {
    const obj: any = {}
    if (message.iceServers?.length) {
      obj.iceServers = message.iceServers.map((e) => IceServerConfig.toJSON(e))
    }
    if (message.iceTransportPolicy !== 0) {
      obj.iceTransportPolicy = iceTransportPolicyToJSON(
        message.iceTransportPolicy,
      )
    }
    if (message.iceCandidatePoolSize !== 0) {
      obj.iceCandidatePoolSize = Math.round(message.iceCandidatePoolSize)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<WebRtcConfig>, I>>(
    base?: I,
  ): WebRtcConfig {
    return WebRtcConfig.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<WebRtcConfig>, I>>(
    object: I,
  ): WebRtcConfig {
    const message = createBaseWebRtcConfig()
    message.iceServers =
      object.iceServers?.map((e) => IceServerConfig.fromPartial(e)) || []
    message.iceTransportPolicy = object.iceTransportPolicy ?? 0
    message.iceCandidatePoolSize = object.iceCandidatePoolSize ?? 0
    return message
  },
}

function createBaseIceServerConfig(): IceServerConfig {
  return { urls: [], username: '', credential: undefined }
}

export const IceServerConfig = {
  encode(
    message: IceServerConfig,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    for (const v of message.urls) {
      writer.uint32(10).string(v!)
    }
    if (message.username !== '') {
      writer.uint32(18).string(message.username)
    }
    switch (message.credential?.$case) {
      case 'password':
        writer.uint32(26).string(message.credential.password)
        break
      case 'oauth':
        IceServerConfig_OauthCredential.encode(
          message.credential.oauth,
          writer.uint32(34).fork(),
        ).ldelim()
        break
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): IceServerConfig {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseIceServerConfig()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.urls.push(reader.string())
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.username = reader.string()
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.credential = { $case: 'password', password: reader.string() }
          continue
        case 4:
          if (tag !== 34) {
            break
          }

          message.credential = {
            $case: 'oauth',
            oauth: IceServerConfig_OauthCredential.decode(
              reader,
              reader.uint32(),
            ),
          }
          continue
      }
      if ((tag & 7) === 4 || tag === 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<IceServerConfig, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<IceServerConfig | IceServerConfig[]>
      | Iterable<IceServerConfig | IceServerConfig[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [IceServerConfig.encode(p).finish()]
        }
      } else {
        yield* [IceServerConfig.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, IceServerConfig>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<IceServerConfig> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [IceServerConfig.decode(p)]
        }
      } else {
        yield* [IceServerConfig.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): IceServerConfig {
    return {
      urls: globalThis.Array.isArray(object?.urls)
        ? object.urls.map((e: any) => globalThis.String(e))
        : [],
      username: isSet(object.username)
        ? globalThis.String(object.username)
        : '',
      credential: isSet(object.password)
        ? { $case: 'password', password: globalThis.String(object.password) }
        : isSet(object.oauth)
          ? {
              $case: 'oauth',
              oauth: IceServerConfig_OauthCredential.fromJSON(object.oauth),
            }
          : undefined,
    }
  },

  toJSON(message: IceServerConfig): unknown {
    const obj: any = {}
    if (message.urls?.length) {
      obj.urls = message.urls
    }
    if (message.username !== '') {
      obj.username = message.username
    }
    if (message.credential?.$case === 'password') {
      obj.password = message.credential.password
    }
    if (message.credential?.$case === 'oauth') {
      obj.oauth = IceServerConfig_OauthCredential.toJSON(
        message.credential.oauth,
      )
    }
    return obj
  },

  create<I extends Exact<DeepPartial<IceServerConfig>, I>>(
    base?: I,
  ): IceServerConfig {
    return IceServerConfig.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<IceServerConfig>, I>>(
    object: I,
  ): IceServerConfig {
    const message = createBaseIceServerConfig()
    message.urls = object.urls?.map((e) => e) || []
    message.username = object.username ?? ''
    if (
      object.credential?.$case === 'password' &&
      object.credential?.password !== undefined &&
      object.credential?.password !== null
    ) {
      message.credential = {
        $case: 'password',
        password: object.credential.password,
      }
    }
    if (
      object.credential?.$case === 'oauth' &&
      object.credential?.oauth !== undefined &&
      object.credential?.oauth !== null
    ) {
      message.credential = {
        $case: 'oauth',
        oauth: IceServerConfig_OauthCredential.fromPartial(
          object.credential.oauth,
        ),
      }
    }
    return message
  },
}

function createBaseIceServerConfig_OauthCredential(): IceServerConfig_OauthCredential {
  return { macKey: '', accessToken: '' }
}

export const IceServerConfig_OauthCredential = {
  encode(
    message: IceServerConfig_OauthCredential,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.macKey !== '') {
      writer.uint32(10).string(message.macKey)
    }
    if (message.accessToken !== '') {
      writer.uint32(18).string(message.accessToken)
    }
    return writer
  },

  decode(
    input: _m0.Reader | Uint8Array,
    length?: number,
  ): IceServerConfig_OauthCredential {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseIceServerConfig_OauthCredential()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.macKey = reader.string()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.accessToken = reader.string()
          continue
      }
      if ((tag & 7) === 4 || tag === 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<IceServerConfig_OauthCredential, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<
          IceServerConfig_OauthCredential | IceServerConfig_OauthCredential[]
        >
      | Iterable<
          IceServerConfig_OauthCredential | IceServerConfig_OauthCredential[]
        >,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [IceServerConfig_OauthCredential.encode(p).finish()]
        }
      } else {
        yield* [IceServerConfig_OauthCredential.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, IceServerConfig_OauthCredential>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<IceServerConfig_OauthCredential> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [IceServerConfig_OauthCredential.decode(p)]
        }
      } else {
        yield* [IceServerConfig_OauthCredential.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): IceServerConfig_OauthCredential {
    return {
      macKey: isSet(object.macKey) ? globalThis.String(object.macKey) : '',
      accessToken: isSet(object.accessToken)
        ? globalThis.String(object.accessToken)
        : '',
    }
  },

  toJSON(message: IceServerConfig_OauthCredential): unknown {
    const obj: any = {}
    if (message.macKey !== '') {
      obj.macKey = message.macKey
    }
    if (message.accessToken !== '') {
      obj.accessToken = message.accessToken
    }
    return obj
  },

  create<I extends Exact<DeepPartial<IceServerConfig_OauthCredential>, I>>(
    base?: I,
  ): IceServerConfig_OauthCredential {
    return IceServerConfig_OauthCredential.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<IceServerConfig_OauthCredential>, I>>(
    object: I,
  ): IceServerConfig_OauthCredential {
    const message = createBaseIceServerConfig_OauthCredential()
    message.macKey = object.macKey ?? ''
    message.accessToken = object.accessToken ?? ''
    return message
  },
}

function createBaseWebRtcSignal(): WebRtcSignal {
  return { body: undefined }
}

export const WebRtcSignal = {
  encode(
    message: WebRtcSignal,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    switch (message.body?.$case) {
      case 'requestOffer':
        writer.uint32(8).uint64(message.body.requestOffer)
        break
      case 'sdp':
        WebRtcSdp.encode(message.body.sdp, writer.uint32(18).fork()).ldelim()
        break
      case 'ice':
        WebRtcIce.encode(message.body.ice, writer.uint32(26).fork()).ldelim()
        break
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): WebRtcSignal {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseWebRtcSignal()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.body = {
            $case: 'requestOffer',
            requestOffer: reader.uint64() as Long,
          }
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.body = {
            $case: 'sdp',
            sdp: WebRtcSdp.decode(reader, reader.uint32()),
          }
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.body = {
            $case: 'ice',
            ice: WebRtcIce.decode(reader, reader.uint32()),
          }
          continue
      }
      if ((tag & 7) === 4 || tag === 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<WebRtcSignal, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<WebRtcSignal | WebRtcSignal[]>
      | Iterable<WebRtcSignal | WebRtcSignal[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [WebRtcSignal.encode(p).finish()]
        }
      } else {
        yield* [WebRtcSignal.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, WebRtcSignal>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<WebRtcSignal> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [WebRtcSignal.decode(p)]
        }
      } else {
        yield* [WebRtcSignal.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): WebRtcSignal {
    return {
      body: isSet(object.requestOffer)
        ? {
            $case: 'requestOffer',
            requestOffer: Long.fromValue(object.requestOffer),
          }
        : isSet(object.sdp)
          ? { $case: 'sdp', sdp: WebRtcSdp.fromJSON(object.sdp) }
          : isSet(object.ice)
            ? { $case: 'ice', ice: WebRtcIce.fromJSON(object.ice) }
            : undefined,
    }
  },

  toJSON(message: WebRtcSignal): unknown {
    const obj: any = {}
    if (message.body?.$case === 'requestOffer') {
      obj.requestOffer = (message.body.requestOffer || Long.UZERO).toString()
    }
    if (message.body?.$case === 'sdp') {
      obj.sdp = WebRtcSdp.toJSON(message.body.sdp)
    }
    if (message.body?.$case === 'ice') {
      obj.ice = WebRtcIce.toJSON(message.body.ice)
    }
    return obj
  },

  create<I extends Exact<DeepPartial<WebRtcSignal>, I>>(
    base?: I,
  ): WebRtcSignal {
    return WebRtcSignal.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<WebRtcSignal>, I>>(
    object: I,
  ): WebRtcSignal {
    const message = createBaseWebRtcSignal()
    if (
      object.body?.$case === 'requestOffer' &&
      object.body?.requestOffer !== undefined &&
      object.body?.requestOffer !== null
    ) {
      message.body = {
        $case: 'requestOffer',
        requestOffer: Long.fromValue(object.body.requestOffer),
      }
    }
    if (
      object.body?.$case === 'sdp' &&
      object.body?.sdp !== undefined &&
      object.body?.sdp !== null
    ) {
      message.body = {
        $case: 'sdp',
        sdp: WebRtcSdp.fromPartial(object.body.sdp),
      }
    }
    if (
      object.body?.$case === 'ice' &&
      object.body?.ice !== undefined &&
      object.body?.ice !== null
    ) {
      message.body = {
        $case: 'ice',
        ice: WebRtcIce.fromPartial(object.body.ice),
      }
    }
    return message
  },
}

function createBaseWebRtcSdp(): WebRtcSdp {
  return { txSeqno: Long.UZERO, sdpType: '', sdp: '' }
}

export const WebRtcSdp = {
  encode(
    message: WebRtcSdp,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (!message.txSeqno.equals(Long.UZERO)) {
      writer.uint32(8).uint64(message.txSeqno)
    }
    if (message.sdpType !== '') {
      writer.uint32(18).string(message.sdpType)
    }
    if (message.sdp !== '') {
      writer.uint32(26).string(message.sdp)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): WebRtcSdp {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseWebRtcSdp()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 8) {
            break
          }

          message.txSeqno = reader.uint64() as Long
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.sdpType = reader.string()
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.sdp = reader.string()
          continue
      }
      if ((tag & 7) === 4 || tag === 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<WebRtcSdp, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<WebRtcSdp | WebRtcSdp[]>
      | Iterable<WebRtcSdp | WebRtcSdp[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [WebRtcSdp.encode(p).finish()]
        }
      } else {
        yield* [WebRtcSdp.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, WebRtcSdp>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<WebRtcSdp> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [WebRtcSdp.decode(p)]
        }
      } else {
        yield* [WebRtcSdp.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): WebRtcSdp {
    return {
      txSeqno: isSet(object.txSeqno)
        ? Long.fromValue(object.txSeqno)
        : Long.UZERO,
      sdpType: isSet(object.sdpType) ? globalThis.String(object.sdpType) : '',
      sdp: isSet(object.sdp) ? globalThis.String(object.sdp) : '',
    }
  },

  toJSON(message: WebRtcSdp): unknown {
    const obj: any = {}
    if (!message.txSeqno.equals(Long.UZERO)) {
      obj.txSeqno = (message.txSeqno || Long.UZERO).toString()
    }
    if (message.sdpType !== '') {
      obj.sdpType = message.sdpType
    }
    if (message.sdp !== '') {
      obj.sdp = message.sdp
    }
    return obj
  },

  create<I extends Exact<DeepPartial<WebRtcSdp>, I>>(base?: I): WebRtcSdp {
    return WebRtcSdp.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<WebRtcSdp>, I>>(
    object: I,
  ): WebRtcSdp {
    const message = createBaseWebRtcSdp()
    message.txSeqno =
      object.txSeqno !== undefined && object.txSeqno !== null
        ? Long.fromValue(object.txSeqno)
        : Long.UZERO
    message.sdpType = object.sdpType ?? ''
    message.sdp = object.sdp ?? ''
    return message
  },
}

function createBaseWebRtcIce(): WebRtcIce {
  return { candidate: '' }
}

export const WebRtcIce = {
  encode(
    message: WebRtcIce,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.candidate !== '') {
      writer.uint32(10).string(message.candidate)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): WebRtcIce {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseWebRtcIce()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.candidate = reader.string()
          continue
      }
      if ((tag & 7) === 4 || tag === 0) {
        break
      }
      reader.skipType(tag & 7)
    }
    return message
  },

  // encodeTransform encodes a source of message objects.
  // Transform<WebRtcIce, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<WebRtcIce | WebRtcIce[]>
      | Iterable<WebRtcIce | WebRtcIce[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [WebRtcIce.encode(p).finish()]
        }
      } else {
        yield* [WebRtcIce.encode(pkt as any).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, WebRtcIce>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<WebRtcIce> {
    for await (const pkt of source) {
      if (globalThis.Array.isArray(pkt)) {
        for (const p of pkt as any) {
          yield* [WebRtcIce.decode(p)]
        }
      } else {
        yield* [WebRtcIce.decode(pkt as any)]
      }
    }
  },

  fromJSON(object: any): WebRtcIce {
    return {
      candidate: isSet(object.candidate)
        ? globalThis.String(object.candidate)
        : '',
    }
  },

  toJSON(message: WebRtcIce): unknown {
    const obj: any = {}
    if (message.candidate !== '') {
      obj.candidate = message.candidate
    }
    return obj
  },

  create<I extends Exact<DeepPartial<WebRtcIce>, I>>(base?: I): WebRtcIce {
    return WebRtcIce.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<WebRtcIce>, I>>(
    object: I,
  ): WebRtcIce {
    const message = createBaseWebRtcIce()
    message.candidate = object.candidate ?? ''
    return message
  },
}

type Builtin =
  | Date
  | Function
  | Uint8Array
  | string
  | number
  | boolean
  | undefined

export type DeepPartial<T> = T extends Builtin
  ? T
  : T extends Long
    ? string | number | Long
    : T extends globalThis.Array<infer U>
      ? globalThis.Array<DeepPartial<U>>
      : T extends ReadonlyArray<infer U>
        ? ReadonlyArray<DeepPartial<U>>
        : T extends { $case: string }
          ? { [K in keyof Omit<T, '$case'>]?: DeepPartial<T[K]> } & {
              $case: T['$case']
            }
          : T extends {}
            ? { [K in keyof T]?: DeepPartial<T[K]> }
            : Partial<T>

type KeysOfUnion<T> = T extends T ? keyof T : never
export type Exact<P, I extends P> = P extends Builtin
  ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & {
      [K in Exclude<keyof I, KeysOfUnion<P>>]: never
    }

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any
  _m0.configure()
}

function isObject(value: any): boolean {
  return typeof value === 'object' && value !== null
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined
}
