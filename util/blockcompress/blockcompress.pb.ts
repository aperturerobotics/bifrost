/* eslint-disable */

export const protobufPackage = 'blockcompress'

/** BlockCompress sets the type of compression to use. */
export enum BlockCompress {
  /** BlockCompress_NONE - BlockCompress_NONE indicates no compression. */
  BlockCompress_NONE = 0,
  /** BlockCompress_SNAPPY - BlockCompress_SNAPPY indicates Snappy compression. */
  BlockCompress_SNAPPY = 1,
  /**
   * BlockCompress_S2 - BlockCompress_S2 indicates S2 compression.
   *
   * S2 is an extension of snappy. S2 is aimed for high throughput, which is why
   * it features concurrent compression for bigger payloads. Decoding is
   * compatible with Snappy compressed content, but content compressed with S2
   * cannot be decompressed by Snappy. This means that S2 can seamlessly replace
   * Snappy without converting compressed content. S2 is designed to have high
   * throughput on content that cannot be compressed. This is important so you
   * don't have to worry about spending CPU cycles on already compressed data.
   *
   * Reference: https://github.com/klauspost/compress/tree/master/s2
   */
  BlockCompress_S2 = 2,
  /** BlockCompress_LZ4 - BlockCompress_LZ4 indicates LZ4 compression. */
  BlockCompress_LZ4 = 3,
  /**
   * BlockCompress_ZSTD - BlockCompress_ZSTD indicates z-standard compression.
   *
   * Zstandard is a real-time compression algorithm, providing high compression
   * ratios. It offers a very wide range of compression / speed trade-off, while
   * being backed by a very fast decoder. A high performance compression
   * algorithm is implemented.
   */
  BlockCompress_ZSTD = 4,
  UNRECOGNIZED = -1,
}

export function blockCompressFromJSON(object: any): BlockCompress {
  switch (object) {
    case 0:
    case 'BlockCompress_NONE':
      return BlockCompress.BlockCompress_NONE
    case 1:
    case 'BlockCompress_SNAPPY':
      return BlockCompress.BlockCompress_SNAPPY
    case 2:
    case 'BlockCompress_S2':
      return BlockCompress.BlockCompress_S2
    case 3:
    case 'BlockCompress_LZ4':
      return BlockCompress.BlockCompress_LZ4
    case 4:
    case 'BlockCompress_ZSTD':
      return BlockCompress.BlockCompress_ZSTD
    case -1:
    case 'UNRECOGNIZED':
    default:
      return BlockCompress.UNRECOGNIZED
  }
}

export function blockCompressToJSON(object: BlockCompress): string {
  switch (object) {
    case BlockCompress.BlockCompress_NONE:
      return 'BlockCompress_NONE'
    case BlockCompress.BlockCompress_SNAPPY:
      return 'BlockCompress_SNAPPY'
    case BlockCompress.BlockCompress_S2:
      return 'BlockCompress_S2'
    case BlockCompress.BlockCompress_LZ4:
      return 'BlockCompress_LZ4'
    case BlockCompress.BlockCompress_ZSTD:
      return 'BlockCompress_ZSTD'
    case BlockCompress.UNRECOGNIZED:
    default:
      return 'UNRECOGNIZED'
  }
}
