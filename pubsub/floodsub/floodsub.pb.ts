// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/pubsub/floodsub/floodsub.proto (package floodsub, syntax proto3)
/* eslint-disable */

import type { HashType } from '../../hash/hash.pb.js'
import { HashType_Enum } from '../../hash/hash.pb.js'
import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import { createMessageType, ScalarType } from '@aptre/protobuf-es-lite'
import { SignedMsg } from '../../peer/peer.pb.js'

export const protobufPackage = 'floodsub'

/**
 * Config configures the floodsub router.
 *
 * @generated from message floodsub.Config
 */
export interface Config {
  /**
   * PublishHashType is the hash type to use when signing published messages.
   * Defaults to sha256
   *
   * @generated from field: hash.HashType publish_hash_type = 1;
   */
  publishHashType?: HashType
}

// Config contains the message type declaration for Config.
export const Config: MessageType<Config> = createMessageType({
  typeName: 'floodsub.Config',
  fields: [
    { no: 1, name: 'publish_hash_type', kind: 'enum', T: HashType_Enum },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})

/**
 * SubscriptionOpts are subscription options.
 *
 * @generated from message floodsub.SubscriptionOpts
 */
export interface SubscriptionOpts {
  /**
   * Subscribe indicates if we are subscribing to this channel ID.
   *
   * @generated from field: bool subscribe = 1;
   */
  subscribe?: boolean
  /**
   * ChannelId is the channel to subscribe to.
   *
   * @generated from field: string channel_id = 2;
   */
  channelId?: string
}

// SubscriptionOpts contains the message type declaration for SubscriptionOpts.
export const SubscriptionOpts: MessageType<SubscriptionOpts> =
  createMessageType({
    typeName: 'floodsub.SubscriptionOpts',
    fields: [
      { no: 1, name: 'subscribe', kind: 'scalar', T: ScalarType.BOOL },
      { no: 2, name: 'channel_id', kind: 'scalar', T: ScalarType.STRING },
    ] as readonly PartialFieldInfo[],
    packedByDefault: true,
  })

/**
 * Packet is the floodsub packet.
 *
 * @generated from message floodsub.Packet
 */
export interface Packet {
  /**
   * Subscriptions contains any new subscription changes.
   *
   * @generated from field: repeated floodsub.SubscriptionOpts subscriptions = 1;
   */
  subscriptions?: SubscriptionOpts[]
  /**
   * Publish contains messages we are publishing.
   *
   * @generated from field: repeated peer.SignedMsg publish = 2;
   */
  publish?: SignedMsg[]
}

// Packet contains the message type declaration for Packet.
export const Packet: MessageType<Packet> = createMessageType({
  typeName: 'floodsub.Packet',
  fields: [
    {
      no: 1,
      name: 'subscriptions',
      kind: 'message',
      T: () => SubscriptionOpts,
      repeated: true,
    },
    {
      no: 2,
      name: 'publish',
      kind: 'message',
      T: () => SignedMsg,
      repeated: true,
    },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
