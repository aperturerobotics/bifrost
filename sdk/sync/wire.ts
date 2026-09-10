import { publicError, SyncError, type SyncErrorCode } from './errors.js'
import { decodeJSON, type JsonValue } from './json.js'
import { ErrorCode, type Failure } from './sync.pb.js'

export const wireVersion = 1

export function toFailure(error: unknown): Failure {
  const safe = publicError(error)
  return {
    code: ErrorCode[safe.code],
    message: safe.message,
    requestId: safe.requestId,
  }
}

export function checkFailure(failure?: Failure): void {
  if (!failure) return
  const code = failure.code ? ErrorCode[failure.code] : undefined
  throw new SyncError(
    (code ?? 'UNAVAILABLE') as SyncErrorCode,
    failure.message || 'The operation failed',
    failure.requestId || undefined,
  )
}

export function readInput(bytes?: Uint8Array): JsonValue {
  try {
    return decodeJSON(bytes ?? new Uint8Array())
  } catch {
    throw new SyncError('VALIDATION', 'Input must be valid UTF-8 JSON')
  }
}
