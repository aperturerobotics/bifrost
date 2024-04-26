// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/hash/hash.proto (package hash, syntax proto3)
/* eslint-disable */

import {
  createEnumType,
  createMessageType,
  Message,
  MessageType,
  PartialFieldInfo,
} from '@aptre/protobuf-es-lite'

export const protobufPackage = 'hash'

/**
 * HashType identifies the hash type in use.
 *
 * @generated from enum hash.HashType
 */
export enum HashType {
  /**
   * HashType_UNKNOWN is an unknown hash type.
   *
   * @generated from enum value: HashType_UNKNOWN = 0;
   */
  HashType_UNKNOWN = 0,

  /**
   * HashType_SHA256 is the sha256 hash type.
   *
   * @generated from enum value: HashType_SHA256 = 1;
   */
  HashType_SHA256 = 1,

  /**
   * HashType_SHA1 is the sha1 hash type.
   * NOTE: Do not use SHA1 unless you absolutely have to for backwards compat! (Git)
   *
   * @generated from enum value: HashType_SHA1 = 2;
   */
  HashType_SHA1 = 2,

  /**
   * HashType_BLAKE3 is the blake3 hash type.
   * Uses a 32-byte digest size.
   *
   * @generated from enum value: HashType_BLAKE3 = 3;
   */
  HashType_BLAKE3 = 3,
}

// HashType_Enum is the enum type for HashType.
export const HashType_Enum = createEnumType('hash.HashType', [
  { no: 0, name: 'HashType_UNKNOWN' },
  { no: 1, name: 'HashType_SHA256' },
  { no: 2, name: 'HashType_SHA1' },
  { no: 3, name: 'HashType_BLAKE3' },
])

/**
 * Hash is a hash of a binary blob.
 *
 * @generated from message hash.Hash
 */
export interface Hash extends Message<Hash> {
  /**
   * HashType is the hash type in use.
   *
   * @generated from field: hash.HashType hash_type = 1;
   */
  hashType?: HashType
  /**
   * Hash is the hash value.
   *
   * @generated from field: bytes hash = 2;
   */
  hash?: Uint8Array
}

export const Hash: MessageType<Hash> = createMessageType({
  typeName: 'hash.Hash',
  fields: [
    { no: 1, name: 'hash_type', kind: 'enum', T: HashType_Enum },
    { no: 2, name: 'hash', kind: 'scalar', T: 12 /* ScalarType.BYTES */ },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
