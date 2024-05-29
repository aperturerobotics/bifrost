// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/transport/common/dialer/dialer.proto (package dialer, syntax proto3)
/* eslint-disable */

import { Backoff } from '@go/github.com/aperturerobotics/util/backoff/backoff.pb.js'
import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import { createMessageType, ScalarType } from '@aptre/protobuf-es-lite'

export const protobufPackage = 'dialer'

/**
 * DialerOpts contains options relating to dialing a statically configured peer.
 *
 * @generated from message dialer.DialerOpts
 */
export interface DialerOpts {
  /**
   * Address is the address of the peer, in the format expected by the transport.
   *
   * @generated from field: string address = 1;
   */
  address?: string
  /**
   * Backoff is the dialing backoff configuration.
   * Can be empty.
   *
   * @generated from field: backoff.Backoff backoff = 2;
   */
  backoff?: Backoff
}

// DialerOpts contains the message type declaration for DialerOpts.
export const DialerOpts: MessageType<DialerOpts> = createMessageType({
  typeName: 'dialer.DialerOpts',
  fields: [
    { no: 1, name: 'address', kind: 'scalar', T: ScalarType.STRING },
    { no: 2, name: 'backoff', kind: 'message', T: () => Backoff },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
