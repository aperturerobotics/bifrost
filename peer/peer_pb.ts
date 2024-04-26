// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/peer/peer.proto (package peer, syntax proto3)
/* eslint-disable */

import {
  createMessageType,
  Message,
  MessageType,
  PartialFieldInfo,
} from '@aptre/protobuf-es-lite'
import type { HashType } from '../hash/hash_pb.js'
import { HashType_Enum } from '../hash/hash_pb.js'

export const protobufPackage = 'peer'

/**
 * Signature contains a signature by a peer.
 *
 * @generated from message peer.Signature
 */
export interface Signature extends Message<Signature> {
  /**
   * PubKey is the public key of the peer.
   * May be empty if the public key is to be inferred from context.
   *
   * @generated from field: bytes pub_key = 1;
   */
  pubKey?: Uint8Array
  /**
   * HashType is the hash type used to hash the data.
   * The signature is then of the hash bytes (usually 32).
   *
   * @generated from field: hash.HashType hash_type = 2;
   */
  hashType?: HashType
  /**
   * SigData contains the signature data.
   * The format is defined by the key type.
   *
   * @generated from field: bytes sig_data = 3;
   */
  sigData?: Uint8Array
}

export const Signature: MessageType<Signature> = createMessageType({
  typeName: 'peer.Signature',
  fields: [
    { no: 1, name: 'pub_key', kind: 'scalar', T: 12 /* ScalarType.BYTES */ },
    { no: 2, name: 'hash_type', kind: 'enum', T: HashType_Enum },
    { no: 3, name: 'sig_data', kind: 'scalar', T: 12 /* ScalarType.BYTES */ },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})

/**
 * SignedMsg is a message from a peer with a signature.
 *
 * @generated from message peer.SignedMsg
 */
export interface SignedMsg extends Message<SignedMsg> {
  /**
   * FromPeerId is the peer identifier of the sender.
   *
   * @generated from field: string from_peer_id = 1;
   */
  fromPeerId?: string
  /**
   * Signature is the sender signature.
   * Should not contain PubKey, which is inferred from peer id.
   *
   * @generated from field: peer.Signature signature = 2;
   */
  signature?: Signature
  /**
   * Data is the signed data.
   *
   * @generated from field: bytes data = 3;
   */
  data?: Uint8Array
}

export const SignedMsg: MessageType<SignedMsg> = createMessageType({
  typeName: 'peer.SignedMsg',
  fields: [
    {
      no: 1,
      name: 'from_peer_id',
      kind: 'scalar',
      T: 9 /* ScalarType.STRING */,
    },
    { no: 2, name: 'signature', kind: 'message', T: Signature },
    { no: 3, name: 'data', kind: 'scalar', T: 12 /* ScalarType.BYTES */ },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
