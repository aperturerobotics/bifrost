/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'bifrost.http.listener'

/**
 * Config configures a http server that listens on a port.
 *
 * Handles incoming requests with LookupHTTPHandler.
 */
export interface Config {
  /**
   * Addr is the address to listen.
   *
   * Example: 0.0.0.0:8080
   */
  addr: string
  /** ClientId is the client id to set on LookupHTTPHandler. */
  clientId: string
  /**
   * CertFile is the path to the certificate file to use for https.
   * Can be unset to use HTTP.
   */
  certFile: string
  /**
   * KeyFile is the path to the key file to use for https.
   * Cannot be unset if cert_file is set.
   * Otherwise can be unset.
   */
  keyFile: string
  /**
   * Wait indicates to wait for LookupHTTPHandler even if it becomes idle.
   * If false: returns 404 not found if LookupHTTPHandler becomes idle.
   */
  wait: boolean
}

function createBaseConfig(): Config {
  return { addr: '', clientId: '', certFile: '', keyFile: '', wait: false }
}

export const Config = {
  encode(
    message: Config,
    writer: _m0.Writer = _m0.Writer.create(),
  ): _m0.Writer {
    if (message.addr !== '') {
      writer.uint32(10).string(message.addr)
    }
    if (message.clientId !== '') {
      writer.uint32(18).string(message.clientId)
    }
    if (message.certFile !== '') {
      writer.uint32(26).string(message.certFile)
    }
    if (message.keyFile !== '') {
      writer.uint32(34).string(message.keyFile)
    }
    if (message.wait === true) {
      writer.uint32(40).bool(message.wait)
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

          message.addr = reader.string()
          continue
        case 2:
          if (tag !== 18) {
            break
          }

          message.clientId = reader.string()
          continue
        case 3:
          if (tag !== 26) {
            break
          }

          message.certFile = reader.string()
          continue
        case 4:
          if (tag !== 34) {
            break
          }

          message.keyFile = reader.string()
          continue
        case 5:
          if (tag !== 40) {
            break
          }

          message.wait = reader.bool()
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
      addr: isSet(object.addr) ? globalThis.String(object.addr) : '',
      clientId: isSet(object.clientId)
        ? globalThis.String(object.clientId)
        : '',
      certFile: isSet(object.certFile)
        ? globalThis.String(object.certFile)
        : '',
      keyFile: isSet(object.keyFile) ? globalThis.String(object.keyFile) : '',
      wait: isSet(object.wait) ? globalThis.Boolean(object.wait) : false,
    }
  },

  toJSON(message: Config): unknown {
    const obj: any = {}
    if (message.addr !== '') {
      obj.addr = message.addr
    }
    if (message.clientId !== '') {
      obj.clientId = message.clientId
    }
    if (message.certFile !== '') {
      obj.certFile = message.certFile
    }
    if (message.keyFile !== '') {
      obj.keyFile = message.keyFile
    }
    if (message.wait === true) {
      obj.wait = message.wait
    }
    return obj
  },

  create<I extends Exact<DeepPartial<Config>, I>>(base?: I): Config {
    return Config.fromPartial(base ?? ({} as any))
  },
  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig()
    message.addr = object.addr ?? ''
    message.clientId = object.clientId ?? ''
    message.certFile = object.certFile ?? ''
    message.keyFile = object.keyFile ?? ''
    message.wait = object.wait ?? false
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

function isSet(value: any): boolean {
  return value !== null && value !== undefined
}
