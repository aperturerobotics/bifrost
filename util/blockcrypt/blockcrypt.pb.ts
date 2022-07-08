/* eslint-disable */
export const protobufPackage = 'blockcrypt'

/** BlockCrypt sets the type of block crypto to use. */
export enum BlockCrypt {
  /** BlockCrypt_UNKNOWN - BlockCrypt_UNKNOWN defaults to BlockCrypt_AES256 */
  BlockCrypt_UNKNOWN = 0,
  /** BlockCrypt_NONE - BlockCrypt_NONE is unencrypted. */
  BlockCrypt_NONE = 1,
  /** BlockCrypt_AES256 - BlockCrypt_AES256 is AES 256-bit block encryption. */
  BlockCrypt_AES256 = 2,
  /** BlockCrypt_AES128 - BlockCrypt_AES128 is AES 128-bit block encryption. */
  BlockCrypt_AES128 = 3,
  /** BlockCrypt_AES192 - BlockCrypt_AES192 is AES 192-bit block encryption. */
  BlockCrypt_AES192 = 4,
  /** BlockCrypt_SM4_16 - BlockCrypt_SM4_16 is SM4 16-bit block encryption. */
  BlockCrypt_SM4_16 = 5,
  /** BlockCrypt_XOR - BlockCrypt_XOR is simple XOR block encryption. */
  BlockCrypt_XOR = 6,
  /** BlockCrypt_3DES - BlockCrypt_3DES is 3des 24-bit block encryption. */
  BlockCrypt_3DES = 7,
  /** BlockCrypt_SALSA20 - BlockCrypt_SALSA20 is salsa20 32-bit block encryption. */
  BlockCrypt_SALSA20 = 8,
  UNRECOGNIZED = -1,
}

export function blockCryptFromJSON(object: any): BlockCrypt {
  switch (object) {
    case 0:
    case 'BlockCrypt_UNKNOWN':
      return BlockCrypt.BlockCrypt_UNKNOWN
    case 1:
    case 'BlockCrypt_NONE':
      return BlockCrypt.BlockCrypt_NONE
    case 2:
    case 'BlockCrypt_AES256':
      return BlockCrypt.BlockCrypt_AES256
    case 3:
    case 'BlockCrypt_AES128':
      return BlockCrypt.BlockCrypt_AES128
    case 4:
    case 'BlockCrypt_AES192':
      return BlockCrypt.BlockCrypt_AES192
    case 5:
    case 'BlockCrypt_SM4_16':
      return BlockCrypt.BlockCrypt_SM4_16
    case 6:
    case 'BlockCrypt_XOR':
      return BlockCrypt.BlockCrypt_XOR
    case 7:
    case 'BlockCrypt_3DES':
      return BlockCrypt.BlockCrypt_3DES
    case 8:
    case 'BlockCrypt_SALSA20':
      return BlockCrypt.BlockCrypt_SALSA20
    case -1:
    case 'UNRECOGNIZED':
    default:
      return BlockCrypt.UNRECOGNIZED
  }
}

export function blockCryptToJSON(object: BlockCrypt): string {
  switch (object) {
    case BlockCrypt.BlockCrypt_UNKNOWN:
      return 'BlockCrypt_UNKNOWN'
    case BlockCrypt.BlockCrypt_NONE:
      return 'BlockCrypt_NONE'
    case BlockCrypt.BlockCrypt_AES256:
      return 'BlockCrypt_AES256'
    case BlockCrypt.BlockCrypt_AES128:
      return 'BlockCrypt_AES128'
    case BlockCrypt.BlockCrypt_AES192:
      return 'BlockCrypt_AES192'
    case BlockCrypt.BlockCrypt_SM4_16:
      return 'BlockCrypt_SM4_16'
    case BlockCrypt.BlockCrypt_XOR:
      return 'BlockCrypt_XOR'
    case BlockCrypt.BlockCrypt_3DES:
      return 'BlockCrypt_3DES'
    case BlockCrypt.BlockCrypt_SALSA20:
      return 'BlockCrypt_SALSA20'
    case BlockCrypt.UNRECOGNIZED:
    default:
      return 'UNRECOGNIZED'
  }
}
