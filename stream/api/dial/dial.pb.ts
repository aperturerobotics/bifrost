/* eslint-disable */
import Long from "long";
import _m0 from "protobufjs/minimal.js";

export const protobufPackage = "stream.api.dial";

/** Config configures the dial controller. */
export interface Config {
  /** PeerId is the remote peer ID to dial. */
  peerId: string;
  /**
   * LocalPeerId is the peer ID to dial with.
   * Can be empty to accept any loaded peer.
   */
  localPeerId: string;
  /** ProtocolId is the protocol ID to dial with. */
  protocolId: string;
  /**
   * TransportId constrains the transport ID to dial with.
   * Can be empty.
   */
  transportId: Long;
  /** Encrypted indicates the stream should be encrypted. */
  encrypted: boolean;
  /** Reliable indicates the stream should be reliable. */
  reliable: boolean;
}

function createBaseConfig(): Config {
  return { peerId: "", localPeerId: "", protocolId: "", transportId: Long.UZERO, encrypted: false, reliable: false };
}

export const Config = {
  encode(message: Config, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    if (message.peerId !== "") {
      writer.uint32(10).string(message.peerId);
    }
    if (message.localPeerId !== "") {
      writer.uint32(18).string(message.localPeerId);
    }
    if (message.protocolId !== "") {
      writer.uint32(26).string(message.protocolId);
    }
    if (!message.transportId.isZero()) {
      writer.uint32(32).uint64(message.transportId);
    }
    if (message.encrypted === true) {
      writer.uint32(40).bool(message.encrypted);
    }
    if (message.reliable === true) {
      writer.uint32(48).bool(message.reliable);
    }
    return writer;
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): Config {
    const reader = input instanceof _m0.Reader ? input : new _m0.Reader(input);
    let end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseConfig();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.peerId = reader.string();
          break;
        case 2:
          message.localPeerId = reader.string();
          break;
        case 3:
          message.protocolId = reader.string();
          break;
        case 4:
          message.transportId = reader.uint64() as Long;
          break;
        case 5:
          message.encrypted = reader.bool();
          break;
        case 6:
          message.reliable = reader.bool();
          break;
        default:
          reader.skipType(tag & 7);
          break;
      }
    }
    return message;
  },

  // encodeTransform encodes a source of message objects.
  // Transform<Config, Uint8Array>
  async *encodeTransform(
    source: AsyncIterable<Config | Config[]> | Iterable<Config | Config[]>,
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Config.encode(p).finish()];
        }
      } else {
        yield* [Config.encode(pkt).finish()];
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, Config>
  async *decodeTransform(
    source: AsyncIterable<Uint8Array | Uint8Array[]> | Iterable<Uint8Array | Uint8Array[]>,
  ): AsyncIterable<Config> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [Config.decode(p)];
        }
      } else {
        yield* [Config.decode(pkt)];
      }
    }
  },

  fromJSON(object: any): Config {
    return {
      peerId: isSet(object.peerId) ? String(object.peerId) : "",
      localPeerId: isSet(object.localPeerId) ? String(object.localPeerId) : "",
      protocolId: isSet(object.protocolId) ? String(object.protocolId) : "",
      transportId: isSet(object.transportId) ? Long.fromValue(object.transportId) : Long.UZERO,
      encrypted: isSet(object.encrypted) ? Boolean(object.encrypted) : false,
      reliable: isSet(object.reliable) ? Boolean(object.reliable) : false,
    };
  },

  toJSON(message: Config): unknown {
    const obj: any = {};
    message.peerId !== undefined && (obj.peerId = message.peerId);
    message.localPeerId !== undefined && (obj.localPeerId = message.localPeerId);
    message.protocolId !== undefined && (obj.protocolId = message.protocolId);
    message.transportId !== undefined && (obj.transportId = (message.transportId || Long.UZERO).toString());
    message.encrypted !== undefined && (obj.encrypted = message.encrypted);
    message.reliable !== undefined && (obj.reliable = message.reliable);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig();
    message.peerId = object.peerId ?? "";
    message.localPeerId = object.localPeerId ?? "";
    message.protocolId = object.protocolId ?? "";
    message.transportId = (object.transportId !== undefined && object.transportId !== null)
      ? Long.fromValue(object.transportId)
      : Long.UZERO;
    message.encrypted = object.encrypted ?? false;
    message.reliable = object.reliable ?? false;
    return message;
  },
};

type Builtin = Date | Function | Uint8Array | string | number | boolean | undefined;

export type DeepPartial<T> = T extends Builtin ? T
  : T extends Long ? string | number | Long : T extends Array<infer U> ? Array<DeepPartial<U>>
  : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>>
  : T extends { $case: string } ? { [K in keyof Omit<T, "$case">]?: DeepPartial<T[K]> } & { $case: T["$case"] }
  : T extends {} ? { [K in keyof T]?: DeepPartial<T[K]> }
  : Partial<T>;

type KeysOfUnion<T> = T extends T ? keyof T : never;
export type Exact<P, I extends P> = P extends Builtin ? P
  : P & { [K in keyof P]: Exact<P[K], I[K]> } & { [K in Exclude<keyof I, KeysOfUnion<P>>]: never };

if (_m0.util.Long !== Long) {
  _m0.util.Long = Long as any;
  _m0.configure();
}

function isSet(value: any): boolean {
  return value !== null && value !== undefined;
}
