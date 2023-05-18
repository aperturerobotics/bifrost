/* eslint-disable */
import Long from 'long'
import _m0 from 'protobufjs/minimal.js'

export const protobufPackage = 'drpc.e2e'

/** MockRequest is the mock request. */
export interface MockRequest {
  /** Body is the body of the request. */
  body: string
}

/** MockResponse is the mock response. */
export interface MockResponse {
  /** ReqBody is the echoed request body. */
  reqBody: string
}

function createBaseMockRequest(): MockRequest {
  return { body: '' }
}

export const MockRequest = {
  encode(
    message: MockRequest,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.body !== '') {
      writer.uint32(10).string(message.body)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): MockRequest {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseMockRequest()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.body = reader.string()
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
  // Transform<MockRequest, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<MockRequest | MockRequest[]>
      | Iterable<MockRequest | MockRequest[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [MockRequest.encode(p).finish()]
        }
      } else {
        yield* [MockRequest.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, MockRequest>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<MockRequest> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [MockRequest.decode(p)]
        }
      } else {
        yield* [MockRequest.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): MockRequest {
    return { body: isSet(object.body) ? String(object.body) : '' }
  },

  toJSON(message: MockRequest): unknown {
    const obj: any = {}
    message.body !== undefined && (obj.body = message.body)
    return obj
  },

  create<I extends Exact<DeepPartial<MockRequest>, I>>(base?: I): MockRequest {
    return MockRequest.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<MockRequest>, I>>(
    object: I
  ): MockRequest {
    const message = createBaseMockRequest()
    message.body = object.body ?? ''
    return message
  },
}

function createBaseMockResponse(): MockResponse {
  return { reqBody: '' }
}

export const MockResponse = {
  encode(
    message: MockResponse,
    writer: _m0.Writer = _m0.Writer.create()
  ): _m0.Writer {
    if (message.reqBody !== '') {
      writer.uint32(10).string(message.reqBody)
    }
    return writer
  },

  decode(input: _m0.Reader | Uint8Array, length?: number): MockResponse {
    const reader =
      input instanceof _m0.Reader ? input : _m0.Reader.create(input)
    let end = length === undefined ? reader.len : reader.pos + length
    const message = createBaseMockResponse()
    while (reader.pos < end) {
      const tag = reader.uint32()
      switch (tag >>> 3) {
        case 1:
          if (tag !== 10) {
            break
          }

          message.reqBody = reader.string()
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
  // Transform<MockResponse, Uint8Array>
  async *encodeTransform(
    source:
      | AsyncIterable<MockResponse | MockResponse[]>
      | Iterable<MockResponse | MockResponse[]>
  ): AsyncIterable<Uint8Array> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [MockResponse.encode(p).finish()]
        }
      } else {
        yield* [MockResponse.encode(pkt).finish()]
      }
    }
  },

  // decodeTransform decodes a source of encoded messages.
  // Transform<Uint8Array, MockResponse>
  async *decodeTransform(
    source:
      | AsyncIterable<Uint8Array | Uint8Array[]>
      | Iterable<Uint8Array | Uint8Array[]>
  ): AsyncIterable<MockResponse> {
    for await (const pkt of source) {
      if (Array.isArray(pkt)) {
        for (const p of pkt) {
          yield* [MockResponse.decode(p)]
        }
      } else {
        yield* [MockResponse.decode(pkt)]
      }
    }
  },

  fromJSON(object: any): MockResponse {
    return { reqBody: isSet(object.reqBody) ? String(object.reqBody) : '' }
  },

  toJSON(message: MockResponse): unknown {
    const obj: any = {}
    message.reqBody !== undefined && (obj.reqBody = message.reqBody)
    return obj
  },

  create<I extends Exact<DeepPartial<MockResponse>, I>>(
    base?: I
  ): MockResponse {
    return MockResponse.fromPartial(base ?? {})
  },

  fromPartial<I extends Exact<DeepPartial<MockResponse>, I>>(
    object: I
  ): MockResponse {
    const message = createBaseMockResponse()
    message.reqBody = object.reqBody ?? ''
    return message
  },
}

/** EndToEnd is a end to end test service. */
export interface EndToEnd {
  /** Mock performs the mock request. */
  Mock(request: MockRequest, abortSignal?: AbortSignal): Promise<MockResponse>
}

export class EndToEndClientImpl implements EndToEnd {
  private readonly rpc: Rpc
  private readonly service: string
  constructor(rpc: Rpc, opts?: { service?: string }) {
    this.service = opts?.service || 'drpc.e2e.EndToEnd'
    this.rpc = rpc
    this.Mock = this.Mock.bind(this)
  }
  Mock(request: MockRequest, abortSignal?: AbortSignal): Promise<MockResponse> {
    const data = MockRequest.encode(request).finish()
    const promise = this.rpc.request(
      this.service,
      'Mock',
      data,
      abortSignal || undefined
    )
    return promise.then((data) => MockResponse.decode(_m0.Reader.create(data)))
  }
}

/** EndToEnd is a end to end test service. */
export type EndToEndDefinition = typeof EndToEndDefinition
export const EndToEndDefinition = {
  name: 'EndToEnd',
  fullName: 'drpc.e2e.EndToEnd',
  methods: {
    /** Mock performs the mock request. */
    mock: {
      name: 'Mock',
      requestType: MockRequest,
      requestStream: false,
      responseType: MockResponse,
      responseStream: false,
      options: {},
    },
  },
} as const

interface Rpc {
  request(
    service: string,
    method: string,
    data: Uint8Array,
    abortSignal?: AbortSignal
  ): Promise<Uint8Array>
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

function isSet(value: any): boolean {
  return value !== null && value !== undefined
}
