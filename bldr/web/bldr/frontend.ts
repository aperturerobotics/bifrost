import { Client, type OpenStreamFunc } from 'starpc'

import {
  Session,
  SendRequest,
  WatchRequest,
} from '../../frontend/frontend.pb.js'
import {
  FrontendClient,
  FrontendDefinition,
} from '../../frontend/frontend_srpc.pb.js'
import { retryWithAbort } from './retry.js'

/** FrontendHandlers is Vite's module runner transport callback boundary. */
interface FrontendHandlers {
  onMessage(payload: unknown): void
}

declare global {
  var __bldrFrontendEnabled: boolean | undefined
  var __bldrFrontend: FrontendResource | undefined
}

/** FrontendResource retains one ordered compiler subscription per document. */
export class FrontendResource {
  private readonly abort = new AbortController()
  private readonly client: FrontendClient
  private readonly ready: Promise<Session>
  private resolveReady!: (session: Session) => void
  private rejectReady!: (reason: unknown) => void
  private session?: Session
  private sequence = 0n
  private handlers?: FrontendHandlers
  private readonly pending: unknown[] = []

  constructor(openStream: OpenStreamFunc) {
    if (globalThis.__bldrFrontend)
      throw new Error('Bldr frontend is already attached to another document')
    globalThis.__bldrFrontend = this
    const rpc = new Client()
    rpc.setOpenStreamFn(openStream)
    this.client = new FrontendClient(rpc, {
      service: `devtool/${FrontendDefinition.typeName}`,
    })
    this.ready = new Promise((resolve, reject) => {
      this.resolveReady = resolve
      this.rejectReady = reject
    })
    void this.ready.catch(() => {})
    void retryWithAbort(
      this.abort.signal,
      async (signal) => {
        let snapshot = true
        for await (const event of this.client.Watch(
          WatchRequest.create(),
          signal,
        )) {
          const sequence = event.sequence ?? 0n
          if (snapshot) {
            snapshot = false
            const session = event.session
            if (
              !session?.id ||
              session.routePrefix !== `/b/fe/${session.id}/`
            ) {
              throw new Error('Bldr frontend returned an invalid session')
            }
            if (
              this.session &&
              (this.session.id !== session.id || this.sequence !== sequence)
            ) {
              this.reload()
              return
            }
            this.session = session
            this.sequence = sequence
            this.resolveReady(session)
            continue
          }
          if (sequence !== this.sequence + 1n) {
            this.reload()
            return
          }
          this.sequence = sequence
          if (event.payload) this.receive(JSON.parse(event.payload))
        }
        throw new Error('Bldr frontend update stream ended')
      },
      {
        errorCb: (error) => console.debug('Bldr frontend reconnecting', error),
      },
    ).catch((error) => {
      if (!this.abort.signal.aborted) this.rejectReady(error)
    })
  }

  /** resolve validates a manifest attachment against the current compiler. */
  public async resolve(entrypoint: string): Promise<string> {
    const session = await this.ready
    if (!session.entrypoints?.includes(entrypoint)) {
      throw new Error(
        `Bldr frontend entrypoint is not configured: ${entrypoint}`,
      )
    }
    return session.routePrefix + entrypoint
  }

  /** connect attaches the upstream client after the initial session snapshot. */
  public async connect(handlers: FrontendHandlers): Promise<void> {
    await this.ready
    this.handlers = handlers
    handlers.onMessage({ type: 'connected' })
    for (const payload of this.pending.splice(0)) handlers.onMessage(payload)
  }

  /** send forwards upstream custom events through the existing RPC client. */
  public async send(payload: unknown): Promise<void> {
    const session = await this.ready
    await this.client.Send(
      SendRequest.create({
        sessionId: session.id,
        payload: JSON.stringify(payload),
      }),
      this.abort.signal,
    )
  }

  /** disconnect detaches Vite; the document still owns the subscription. */
  public disconnect(): void {
    this.handlers = undefined
  }

  /** release cancels the stream and any delayed reconnect. */
  public release(): void {
    this.abort.abort()
    this.rejectReady(new Error('Bldr frontend document closed'))
    this.handlers = undefined
    this.pending.length = 0
    if (globalThis.__bldrFrontend === this)
      globalThis.__bldrFrontend = undefined
  }

  private receive(payload: unknown): void {
    if (this.handlers) this.handlers.onMessage(payload)
    else if (this.pending.length < 64) this.pending.push(payload)
    else this.reload()
  }

  private reload(): void {
    this.release()
    location.reload()
  }
}
