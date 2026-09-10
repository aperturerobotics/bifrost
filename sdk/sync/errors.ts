export type SyncErrorCode =
  | 'VALIDATION'
  | 'AUTHENTICATION'
  | 'DENIED'
  | 'MISSING_COLLECTION'
  | 'SCHEMA_MISMATCH'
  | 'QUERY_LIMIT'
  | 'CONFLICT'
  | 'UNAVAILABLE'
  | 'UNCERTAIN'
  | 'STORAGE'
  | 'CLOSED'

// SyncError carries a stable code and the recovery identity of an uncertain write.
export class SyncError extends Error {
  override readonly name = 'SyncError'

  constructor(
    readonly code: SyncErrorCode,
    message: string,
    readonly requestId?: string,
  ) {
    super(message)
  }
}

// publicError keeps internal failures out of application and transport responses.
export function publicError(error: unknown): SyncError {
  if (error instanceof SyncError) return error
  return new SyncError('STORAGE', 'The data operation could not complete')
}
