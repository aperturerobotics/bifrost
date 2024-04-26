// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/signaling/rpc/signaling.proto (package signaling.rpc, syntax proto3)
/* eslint-disable */

import {
  createMessageType,
  Message,
  MessageType,
  PartialFieldInfo,
} from '@aptre/protobuf-es-lite'
import { SignedMsg } from '../../peer/peer_pb.js'

export const protobufPackage = 'signaling.rpc'

/**
 * ListenRequest is the body of the Listen request.
 *
 * @generated from message signaling.rpc.ListenRequest
 */
export interface ListenRequest extends Message<ListenRequest> {}

export const ListenRequest: MessageType<ListenRequest> = createMessageType({
  typeName: 'signaling.rpc.ListenRequest',
  fields: [] as readonly PartialFieldInfo[],
  packedByDefault: true,
})

/**
 * ListenResponse is a message sent in a stream in response to Listen.
 *
 * @generated from message signaling.rpc.ListenResponse
 */
export interface ListenResponse extends Message<ListenResponse> {
  /**
   * Body is the body of the response.
   *
   * @generated from oneof signaling.rpc.ListenResponse.body
   */
  body?:
    | {
        value?: undefined
        case: undefined
      }
    | {
        /**
         * SetPeer marks that a remote peer wants a session.
         * The contents of the string is the encoded peer id of the remote peer.
         *
         * @generated from field: string set_peer = 1;
         */
        value: string
        case: 'setPeer'
      }
    | {
        /**
         * ClearPeer marks that a remote peer no longer wants a session.
         * The contents of the string is the encoded peer id of the remote peer.
         *
         * @generated from field: string clear_peer = 2;
         */
        value: string
        case: 'clearPeer'
      }
}

export const ListenResponse: MessageType<ListenResponse> = createMessageType({
  typeName: 'signaling.rpc.ListenResponse',
  fields: [
    {
      no: 1,
      name: 'set_peer',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
      oneof: 'body',
    },
    {
      no: 2,
      name: 'clear_peer',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
      oneof: 'body',
    },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})

/**
 * SessionInit is a message to init a Session.
 *
 * @generated from message signaling.rpc.SessionInit
 */
export interface SessionInit extends Message<SessionInit> {
  /**
   * PeerId is the remote peer id we want to contact.
   *
   * @generated from field: string peer_id = 1;
   */
  peerId?: string
}

export const SessionInit: MessageType<SessionInit> = createMessageType({
  typeName: 'signaling.rpc.SessionInit',
  fields: [
    { no: 1, name: 'peer_id', kind: 'scalar', T: 9 /* ScalarType.STRING */ },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})

/**
 * SessionMsg contains a signed message and a sequence number.
 *
 * @generated from message signaling.rpc.SessionMsg
 */
export interface SessionMsg extends Message<SessionMsg> {
  /**
   * SignedMsg is the signed message body.
   *
   * @generated from field: peer.SignedMsg signed_msg = 1;
   */
  signedMsg?: SignedMsg
  /**
   * Seqno is the message sequence number for clear and ack.
   *
   * @generated from field: uint64 seqno = 2;
   */
  seqno?: bigint
}

export const SessionMsg: MessageType<SessionMsg> = createMessageType({
  typeName: 'signaling.rpc.SessionMsg',
  fields: [
    { no: 1, name: 'signed_msg', kind: 'message', T: SignedMsg },
    { no: 2, name: 'seqno', kind: 'scalar', T: 4 /* ScalarType.UINT64 */ },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})

/**
 * SessionRequest is a message sent from the client to the server.
 *
 * @generated from message signaling.rpc.SessionRequest
 */
export interface SessionRequest extends Message<SessionRequest> {
  /**
   * SessionSeqno is the session sequence number.
   * If this doesn't match the current session no, this pkt will be dropped.
   * This should be zero for the init packet.
   *
   * @generated from field: uint64 session_seqno = 1;
   */
  sessionSeqno?: bigint

  /**
   * Body is the body of the request.
   *
   * @generated from oneof signaling.rpc.SessionRequest.body
   */
  body?:
    | {
        value?: undefined
        case: undefined
      }
    | {
        /**
         * Init initializes the session setting which peer to contact.
         *
         * @generated from field: signaling.rpc.SessionInit init = 2;
         */
        value: SessionInit
        case: 'init'
      }
    | {
        /**
         * SendMsg sends a signed message to the remote peer.
         * The server will buffer at most one message for the remote peer at a time.
         * If there is an existing pending outgoing message this will overwrite it.
         * Wait for the message received ack before sending again to avoid overwriting.
         * The signature must match the peer id associated with the rpc session.
         *
         * @generated from field: signaling.rpc.SessionMsg send_msg = 3;
         */
        value: SessionMsg
        case: 'sendMsg'
      }
    | {
        /**
         * ClearMsg clears a previously sent message from the outbox.
         * If the sequence number does not match, does nothing.
         *
         * @generated from field: uint64 clear_msg = 4;
         */
        value: bigint
        case: 'clearMsg'
      }
    | {
        /**
         * AckMsg acknowledges that the current incoming message was processed.
         * If the id doesn't match the current incoming message seqno, does nothing.
         *
         * @generated from field: uint64 ack_msg = 5;
         */
        value: bigint
        case: 'ackMsg'
      }
}

export const SessionRequest: MessageType<SessionRequest> = createMessageType({
  typeName: 'signaling.rpc.SessionRequest',
  fields: [
    {
      no: 1,
      name: 'session_seqno',
      kind: 'scalar',
      T: 4 /* ScalarType.UINT64 */,
    },
    { no: 2, name: 'init', kind: 'message', T: SessionInit, oneof: 'body' },
    { no: 3, name: 'send_msg', kind: 'message', T: SessionMsg, oneof: 'body' },
    {
      no: 4,
      name: 'clear_msg',
      kind: 'scalar',
      T: 4 /* ScalarType.UINT64 */,
      oneof: 'body',
    },
    {
      no: 5,
      name: 'ack_msg',
      kind: 'scalar',
      T: 4 /* ScalarType.UINT64 */,
      oneof: 'body',
    },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})

/**
 * SessionResponse is a message sent from the server to the client.
 *
 * @generated from message signaling.rpc.SessionResponse
 */
export interface SessionResponse extends Message<SessionResponse> {
  /**
   * Body is the body of the request.
   *
   * @generated from oneof signaling.rpc.SessionResponse.body
   */
  body?:
    | {
        value?: undefined
        case: undefined
      }
    | {
        /**
         * Opened indicates the connection w/ the remote peer is open.
         * Contains the session nonce, incremented if the peer re-connected.
         * If this increments, re-send the outgoing message.
         *
         * @generated from field: uint64 opened = 1;
         */
        value: bigint
        case: 'opened'
      }
    | {
        /**
         * Closed indicates the connection w/ the remote peer is closed.
         * This is the assumed initial state.
         *
         * @generated from field: bool closed = 2;
         */
        value: boolean
        case: 'closed'
      }
    | {
        /**
         * RecvMsg transfers a message sent by the remote peer.
         * Send ACK once the message has been processed.
         *
         * @generated from field: signaling.rpc.SessionMsg recv_msg = 3;
         */
        value: SessionMsg
        case: 'recvMsg'
      }
    | {
        /**
         * ClearMsg clears a message from the remote peer previously sent with recv_msg.
         * This is sent if the remote peer clears their outgoing message before acked.
         *
         * @generated from field: uint64 clear_msg = 4;
         */
        value: bigint
        case: 'clearMsg'
      }
    | {
        /**
         * AckMsg confirms that the outgoing message was sent and acked.
         *
         * @generated from field: uint64 ack_msg = 5;
         */
        value: bigint
        case: 'ackMsg'
      }
}

export const SessionResponse: MessageType<SessionResponse> = createMessageType({
  typeName: 'signaling.rpc.SessionResponse',
  fields: [
    {
      no: 1,
      name: 'opened',
      kind: 'scalar',
      T: 4 /* ScalarType.UINT64 */,
      oneof: 'body',
    },
    {
      no: 2,
      name: 'closed',
      kind: 'scalar',
      T: 8 /* ScalarType.BOOL */,
      oneof: 'body',
    },
    { no: 3, name: 'recv_msg', kind: 'message', T: SessionMsg, oneof: 'body' },
    {
      no: 4,
      name: 'clear_msg',
      kind: 'scalar',
      T: 4 /* ScalarType.UINT64 */,
      oneof: 'body',
    },
    {
      no: 5,
      name: 'ack_msg',
      kind: 'scalar',
      T: 4 /* ScalarType.UINT64 */,
      oneof: 'body',
    },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
