/* eslint-disable */
import { Backoff } from "@go/github.com/aperturerobotics/util/backoff/backoff.pb.js";
import Long from "long";
import _m0 from "protobufjs/minimal.js";

export const protobufPackage = "stream.srpc.client";

/** Config configures a client for a srpc service. */
export interface Config {
  /** ServerPeerIds are the static list of peer IDs to contact. */
  serverPeerIds: string[];
  /**
   * PerServerBackoff is the server peer error backoff configuration.
   * Can be empty.
   */
  perServerBackoff:
    | Backoff
    | undefined;
  /**
   * SrcPeerId is the source peer id to contact from.
   * Can be empty.
   */
  srcPeerId: string;
  /** TransportId restricts which transport we can dial out from. */
  transportId: Long;
  /**
   * TimeoutDur sets the per-server establish timeout.
   * If unset, no timeout.
   * Example: 15s
   */
  timeoutDur: string;
}

function createBaseConfig(): Config {
  return { serverPeerIds: [], perServerBackoff: undefined, srcPeerId: "", transportId: Long.UZERO, timeoutDur: "" };
}

export const Config = {
  encode(message: Config, writer: _m0.Writer = _m0.Writer.create()): _m0.Writer {
    for (const v of message.serverPeerIds) {
      writer.uint32(10).string(v!);
    }
    if (message.perServerBackoff !== undefined) {
      Backoff.encode(message.perServerBackoff, writer.uint32(18).fork()).ldelim();
    }
    if (message.srcPeerId !== "") {
      writer.uint32(26).string(message.srcPeerId);
    }
    if (!message.transportId.isZero()) {
      writer.uint32(32).uint64(message.transportId);
    }
    if (message.timeoutDur !== "") {
      writer.uint32(42).string(message.timeoutDur);
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
          message.serverPeerIds.push(reader.string());
          break;
        case 2:
          message.perServerBackoff = Backoff.decode(reader, reader.uint32());
          break;
        case 3:
          message.srcPeerId = reader.string();
          break;
        case 4:
          message.transportId = reader.uint64() as Long;
          break;
        case 5:
          message.timeoutDur = reader.string();
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
      serverPeerIds: Array.isArray(object?.serverPeerIds) ? object.serverPeerIds.map((e: any) => String(e)) : [],
      perServerBackoff: isSet(object.perServerBackoff) ? Backoff.fromJSON(object.perServerBackoff) : undefined,
      srcPeerId: isSet(object.srcPeerId) ? String(object.srcPeerId) : "",
      transportId: isSet(object.transportId) ? Long.fromValue(object.transportId) : Long.UZERO,
      timeoutDur: isSet(object.timeoutDur) ? String(object.timeoutDur) : "",
    };
  },

  toJSON(message: Config): unknown {
    const obj: any = {};
    if (message.serverPeerIds) {
      obj.serverPeerIds = message.serverPeerIds.map((e) => e);
    } else {
      obj.serverPeerIds = [];
    }
    message.perServerBackoff !== undefined &&
      (obj.perServerBackoff = message.perServerBackoff ? Backoff.toJSON(message.perServerBackoff) : undefined);
    message.srcPeerId !== undefined && (obj.srcPeerId = message.srcPeerId);
    message.transportId !== undefined && (obj.transportId = (message.transportId || Long.UZERO).toString());
    message.timeoutDur !== undefined && (obj.timeoutDur = message.timeoutDur);
    return obj;
  },

  fromPartial<I extends Exact<DeepPartial<Config>, I>>(object: I): Config {
    const message = createBaseConfig();
    message.serverPeerIds = object.serverPeerIds?.map((e) => e) || [];
    message.perServerBackoff = (object.perServerBackoff !== undefined && object.perServerBackoff !== null)
      ? Backoff.fromPartial(object.perServerBackoff)
      : undefined;
    message.srcPeerId = object.srcPeerId ?? "";
    message.transportId = (object.transportId !== undefined && object.transportId !== null)
      ? Long.fromValue(object.transportId)
      : Long.UZERO;
    message.timeoutDur = object.timeoutDur ?? "";
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
