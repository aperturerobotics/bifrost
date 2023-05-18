/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'
import { DialerOpts } from '../common/dialer/dialer.pb.js'
import { Opts } from '../common/quic/quic.pb.js'

export const protobufPackage = 'websocket'

/**
 * Config is the configuration for the Websocket transport.
 *
 * Bifrost speaks Quic over the websocket. While this is not always necessary,
 * especially when using wss transports, we still need to ensure end-to-end
 * encryption to the peer that we handshake with on the other end, and to manage
 * stream congestion control, multiplexing,
 */
export interface Config {
  /**
   * TransportPeerID sets the peer ID to attach the transport to.
   * If unset, attaches to any running peer with a private key.
   */
  transportPeerId: string
  /**
   * ListenAddr contains the address to listen on.
   * Has no effect in the browser.
   */
  listenAddr: string
  /**
   * QuicOpts are the quic protocol options.
   *
   * The WebSocket transport always disables FEC and several other UDP-centric
   * features which are unnecessary due to the "reliable" nature of WebSockets.
   */
  quic: Opts | undefined
  /** Dialers maps peer IDs to dialers. */
  dialers: { [key: string]: DialerOpts }
  /**
   * RestrictPeerId restricts all incoming peer IDs to the given ID.
   * Any other peers trying to connect will be disconneted at handshake time.
   */
  restrictPeerId: string
}

export interface Config_DialersEntry {
  key: string
  value: DialerOpts | undefined
}

function createBaseConfig(): Config {
  return {
    transportPeerId: '',
    listenAddr: '',
    quic: undefined,
    dialers: {},
    restrictPeerId: '',
  }
}

export const Config = {
  encode(
    message: Config,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.transportPeerId !== '') {
      writer.uint32(10).string(message.transportPeerId)
    }
    if (message.listenAddr !== '') {
      writer.uint32(18).string(message.listenAddr)
    }
    if (message.quic !== undefined) {
      Opts.encode(message.quic, writer.uint32(26).fork()).ldelim()
    }
    Object.entries(message.dialers).forEach(([key, value]) => {
      Config_DialersEntry.encode(
        { key: key as any, value },
        writer.uint32(34).fork()
      ).ldelim()
    })
    if (message.restrictPeerId !== '') {
      writer.uint32(42).string(message.restrictPeerId)
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

          message.transportPeerId = reader.string()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.listenAddr = reader.string()
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.quic = Opts.decode(reader, reader.uint32())
          continue
        case 4:
          if (tag !== 34) {
            break
          }

          const entry4 = Config_DialersEntry.decode(reader, reader.uint32())
          if (entry4.value !== undefined) {
            message.dialers[entry4.key] = entry4.value
          }
          continue
        case 5:
          if (tag !== 42) {
            break
          }

          message.restrictPeerId = reader.string()
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
    source: AsyncIterable<Config | Config[]> | Iterable<Config | Config[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Config.encode(p).finish()]
        }
      } else {
        yield* [Config.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Config>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<Config> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Config.decode(p)]
        }
      } else {
        yield* [Config.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Config {
    return {
      transportPeerId: isSet(object.transportPeerId)
        ? String(object.transportPeerId)
        : '',
      listenAddr: isSet(object.listenAddr) ? String(object.listenAddr) : '',
      quic: isSet(object.quic) ? Opts.fromJSON(object.quic) : undefined,
      dialers: isObject(object.dialers)
        ? Object.entries(object.dialers).reduce<{ [key: string]: DialerOpts }>(
            (acc, [key, value]) => {
              acc[key] = DialerOpts.fromJSON(value)
              return acc
            },
            {}
          )
        : {},
      restrictPeerId: isSet(object.restrictPeerId)
        ? String(object.restrictPeerId)
        : '',
    }
  },

  toJSON(message: Config): unknown {
    const obj: any = {}
    message.transportPeerId !== undefined &&
      (obj.transportPeerId = message.transportPeerId)
    message.listenAddr !== undefined && (obj.listenAddr = message.listenAddr)
    message.quic !== undefined &&
      (obj.quic = message.quic ? Opts.toJSON(message.quic) : undefined)
    obj.dialers = {}
    if (message.dialers) {
      Object.entries(message.dialers).forEach(([k, v]) => {
        obj.dialers[k] = DialerOpts.toJSON(v)
      })
    }
    message.restrictPeerId !== undefined &&
      (obj.restrictPeerId = message.restrictPeerId)
    return obj
  },

  create<I extends Exact<DeepPartial<Config>, I>>(base?: I): Config {
    return Config.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig()
    message.transportPeerId = object.transportPeerId ?? ''
    message.listenAddr = object.listenAddr ?? ''
    message.quic =
      object.quic !== undefined && object.quic !== null
        ? Opts.fromPartial(object.quic)
        : undefined
    message.dialers = Object.entries(object.dialers ?? {}).reduce<{
      [key: string]: DialerOpts
    }>((acc, [key, value]) => {
      if (value !== undefined) {
        acc[key] = DialerOpts.fromPartial(value)
      }
      return acc
    }, {})
    message.restrictPeerId = object.restrictPeerId ?? ''
    return message
  },
}

function createBaseConfig_DialersEntry(): Config_DialersEntry {
  return { key: '', value: undefined }
}

export const Config_DialersEntry = {
  encode(
    message: Config_DialersEntry,
    writer: _m0.Writer = _m0.Writer.create()
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
      | Iterable<Config_DialersEntry | Config_DialersEntry[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Config_DialersEntry.encode(p).finish()]
        }
      } else {
        yield* [Config_DialersEntry.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Config_DialersEntry>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<Config_DialersEntry> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Config_DialersEntry.decode(p)]
        }
      } else {
        yield* [Config_DialersEntry.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): Config_DialersEntry {
    return {
      key: isSet(object.key) ? String(object.key) : '',
      value: isSet(object.value)
        ? DialerOpts.fromJSON(object.value)
        : undefined,
    }
  },

  toJSON(message: Config_DialersEntry): unknown {
    const obj: any = {}
    message.key !== undefined && (obj.key = message.key)
    message.value !== undefined &&
      (obj.value = message.value ? DialerOpts.toJSON(message.value) : undefined)
    return obj
  },

  create<I extends Exact<DeepPartial<Config_DialersEntry>, I>>(
    base?: I
  ): Config_DialersEntry {
    return Config_DialersEntry.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<Config_DialersEntry>, I>>(
    object: I
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
  : T extends Array<infer U>
  ? Array<DeepPartial<U>>
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
