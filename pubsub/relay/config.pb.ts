// @generated by protoc-gen-es-lite unknown with parameter "target=ts,ts_nocheck=false"
// @generated from file github.com/aperturerobotics/bifrost/pubsub/relay/config.proto (package pubsub.relay, syntax proto3)
/* eslint-disable */

import type { MessageType, PartialFieldInfo } from '@aptre/protobuf-es-lite'
import { createMessageType, Message, ScalarType } from '@aptre/protobuf-es-lite'

export const protobufPackage = 'pubsub.relay'

/**
 * Config is the pubsub relay configuration.
 * The relay controller subscribes to a pubsub topic to ensure the peer relays messages.
 *
 * @generated from message pubsub.relay.Config
 */
export type Config = Message<{
  /**
   * PeerId is the peer ID to look up and use private key for.
   *
   * @generated from field: string peer_id = 1;
   */
  peerId?: string
  /**
   * TopicIds are the list of topic IDs to subscribe to.
   *
   * @generated from field: repeated string topic_ids = 2;
   */
  topicIds?: string[]
}>

// Config contains the message type declaration for Config.
export const Config: MessageType<Config> = createMessageType({
  typeName: 'pubsub.relay.Config',
  fields: [
    { no: 1, name: 'peer_id', kind: 'scalar', T: ScalarType.STRING },
    {
      no: 2,
      name: 'topic_ids',
      kind: 'scalar',
      T: ScalarType.STRING,
      repeated: true,
    },
  ] as readonly PartialFieldInfo[],
  packedByDefault: true,
})
