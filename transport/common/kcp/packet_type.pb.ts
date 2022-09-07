/* eslint-disable */

export const protobufPackage = "kcp";

/** PacketType is a one-byte trailer indicating the type of packet. */
export enum PacketType {
  PacketType_HANDSHAKE = 0,
  PacketType_RAW = 1,
  PacketType_KCP_SMUX = 2,
  PacketType_CLOSE_LINK = 3,
  UNRECOGNIZED = -1,
}

export function packetTypeFromJSON(object: any): PacketType {
  switch (object) {
    case 0:
    case "PacketType_HANDSHAKE":
      return PacketType.PacketType_HANDSHAKE;
    case 1:
    case "PacketType_RAW":
      return PacketType.PacketType_RAW;
    case 2:
    case "PacketType_KCP_SMUX":
      return PacketType.PacketType_KCP_SMUX;
    case 3:
    case "PacketType_CLOSE_LINK":
      return PacketType.PacketType_CLOSE_LINK;
    case -1:
    case "UNRECOGNIZED":
    default:
      return PacketType.UNRECOGNIZED;
  }
}

export function packetTypeToJSON(object: PacketType): string {
  switch (object) {
    case PacketType.PacketType_HANDSHAKE:
      return "PacketType_HANDSHAKE";
    case PacketType.PacketType_RAW:
      return "PacketType_RAW";
    case PacketType.PacketType_KCP_SMUX:
      return "PacketType_KCP_SMUX";
    case PacketType.PacketType_CLOSE_LINK:
      return "PacketType_CLOSE_LINK";
    case PacketType.UNRECOGNIZED:
    default:
      return "UNRECOGNIZED";
  }
}
