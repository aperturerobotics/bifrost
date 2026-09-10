import type { Engine } from '../../sdk/world/engine.js'
import { createAccess, type DatabaseAccess } from '../../sdk/sync/access.js'
import { publicError, SyncError } from '../../sdk/sync/errors.js'
import type { Principal, Schema } from '../../sdk/sync/schema.js'
import {
  Application,
  type ApplicationOptions,
  type Migration,
} from './application.js'
import { openNodeEngine, type OwnedEngine } from './node/host.js'

export type ServerOptions<S extends Schema, P extends Principal> = Omit<
  ApplicationOptions<S, P>,
  'engine'
> &
  (
    | { directory: string; engine?: never }
    | { engine: Engine; directory?: never }
  ) & {
    authenticate(token: string, signal: AbortSignal): Promise<P>
    migration?: Migration<S>
  }

// SyncServer owns admission and any engine it opened; supplied engines retain their owner.
export class SyncServer<
  S extends Schema,
  P extends Principal,
> implements AsyncDisposable {
  private closing?: Promise<void>

  constructor(
    readonly application: Application<S, P>,
    readonly options: ServerOptions<S, P>,
    private readonly owner?: OwnedEngine,
  ) {}

  as(principal: P): DatabaseAccess<S> {
    const identity = { ...principal }
    return createAccess((operation, options) =>
      this.application.execute(identity, operation, options),
    )
  }

  admin(scope: string): DatabaseAccess<S> {
    const principal = { subject: 'admin', scope } as P
    return createAccess((operation, options) =>
      this.application.execute(principal, operation, options, true),
    )
  }

  close(): Promise<void> {
    this.closing ??= (async () => {
      await this.application.close()
      await this.owner?.close()
    })()
    return this.closing
  }

  [Symbol.asyncDispose](): Promise<void> {
    return this.close()
  }
}

// createServer opens and checks storage before returning an admission capability.
export async function createServer<S extends Schema, P extends Principal>(
  options: ServerOptions<S, P>,
): Promise<SyncServer<S, P>> {
  let owner: OwnedEngine | undefined
  let application: Application<S, P> | undefined
  try {
    if (options.directory !== undefined)
      owner = await openNodeEngine(options.directory)
    const engine = options.engine ?? owner?.engine
    if (!engine)
      throw new SyncError(
        'VALIDATION',
        'A directory or supplied engine is required',
      )
    application = new Application({ ...options, engine })
    await application.initialize(options.migration)
    return new SyncServer(application, options, owner)
  } catch (error) {
    await application?.close()
    await owner?.close().catch(() => {})
    throw publicError(error)
  }
}
