// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/stream/forwarding/forwarding.proto (package stream.forwarding, syntax proto3)
/* eslint-disable */

import type { MessageType, PartialFieldInfo } from "@aptre/protobuf-es-lite";
import { createMessageType, Message } from "@aptre/protobuf-es-lite";

export const protobufPackage = "stream.forwarding";

/**
 * Config configures the forwarding controller.
 *
 * @generated from message stream.forwarding.Config
 */
export type Config = Message<{
  /**
   * PeerId is the peer ID to forward for.
   * Can be empty.
   *
   * @generated from field: string peer_id = 1;
   */
  peerId?: string;
  /**
   * ProtocolId is the protocol ID to forward for.
   *
   * @generated from field: string protocol_id = 2;
   */
  protocolId?: string;
  /**
   * TargetMultiaddr is the target multiaddress to dial.
   *
   * @generated from field: string target_multiaddr = 3;
   */
  targetMultiaddr?: string;

}>;

export const Config: MessageType<Config> = createMessageType(
  {
    typeName: "stream.forwarding.Config",
    fields: [
        { no: 1, name: "peer_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
        { no: 2, name: "protocol_id", kind: "scalar", T: 9 /* ScalarType.STRING */ },
        { no: 3, name: "target_multiaddr", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    ] as readonly PartialFieldInfo[],
    packedByDefault: true,
  },
);

