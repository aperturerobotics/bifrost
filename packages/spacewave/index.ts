export { connect } from '../../sdk/sync/client.js'
export type {
  Database,
  Collection,
  ConnectOptions,
  ConnectionState,
  ConnectionStatus,
  SubscriptionState,
  Observer,
} from '../../sdk/sync/client.js'
export { defineSchema } from '../../sdk/sync/schema.js'
export type {
  Schema,
  Principal,
  MutationContext,
  MutationHandlers,
  Transaction,
  RecordEntry,
  Query,
  CallOptions,
  Input,
  Output,
} from '../../sdk/sync/schema.js'
export { SyncError } from '../../sdk/sync/errors.js'
export type { SyncErrorCode } from '../../sdk/sync/errors.js'
export type { JsonValue } from '../../sdk/sync/json.js'
