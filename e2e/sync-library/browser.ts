import { connect, SyncError } from 'spacewave'
import type { Database, SubscriptionState } from 'spacewave'
import { schema } from './schema.js'

// The test transport can discard a reply after durable acceptance without changing RPCs.
const NativeWebSocket = globalThis.WebSocket
const sockets: WebSocket[] = []
let discardReplies = false
class TestWebSocket extends NativeWebSocket {
  private readonly listeners = new Map<
    EventListenerOrEventListenerObject,
    EventListener
  >()

  constructor(url: string | URL, protocols?: string | string[]) {
    super(url, protocols)
    sockets.push(this)
  }

  override addEventListener(
    type: string,
    callback: EventListenerOrEventListenerObject | null,
    options?: AddEventListenerOptions | boolean,
  ): void {
    if (type !== 'message' || !callback) {
      super.addEventListener(type, callback, options)
      return
    }
    const wrapped = (event: Event) => {
      if (discardReplies) return
      if (typeof callback === 'function') callback.call(this, event)
      else callback.handleEvent(event)
    }
    this.listeners.set(callback, wrapped)
    super.addEventListener(type, wrapped, options)
  }

  override removeEventListener(
    type: string,
    callback: EventListenerOrEventListenerObject | null,
    options?: EventListenerOptions | boolean,
  ): void {
    super.removeEventListener(
      type,
      (callback && this.listeners.get(callback)) || callback,
      options,
    )
    if (callback) this.listeners.delete(callback)
  }
}
globalThis.WebSocket = TestWebSocket

let db: Database<typeof schema>
let pending: Promise<unknown> | undefined
const subscriptions = new Map<string, () => void>()
const snapshots: Record<string, SubscriptionState<unknown>[]> = {}
const connections: string[] = []
let tokens = 0

async function outcome(
  call: () => Promise<unknown>,
): Promise<{ result?: unknown; code?: string }> {
  try {
    return { result: await call() }
  } catch (error) {
    return { code: error instanceof SyncError ? error.code : String(error) }
  }
}

export const fixture = {
  async start(url: string, token: string) {
    return outcome(async () => {
      db = await connect({
        url,
        schema,
        getAccessToken: () => {
          tokens++
          return token
        },
      })
      db.connection.subscribe((state) => connections.push(state.status))
    })
  },
  put: (key: string, value: unknown, requestId?: string) =>
    outcome(() =>
      db.collection('todos').put(key, value as never, { requestId }),
    ),
  get: (key: string) => outcome(() => db.collection('todos').get(key)),
  scan: (collection: 'todos' | 'secrets' = 'todos') =>
    outcome(() => db.collection(collection).scan()),
  watch(id: string, prefix = '') {
    snapshots[id] = []
    subscriptions.set(
      id,
      db.collection('todos').subscribe(
        { prefix },
        {
          next: (state) => {
            snapshots[id].push(state)
          },
        },
      ),
    )
  },
  stop(id: string) {
    subscriptions.get(id)?.()
    subscriptions.delete(id)
  },
  latest(id: string) {
    return snapshots[id]?.at(-1)
  },
  states(id: string) {
    return snapshots[id]?.map((state) => state.status)
  },
  status() {
    return {
      state: db.connection.current,
      connections,
      tokens,
      openSockets: sockets.filter(
        (socket) => socket.readyState !== WebSocket.CLOSED,
      ).length,
    }
  },
  disconnect() {
    sockets.at(-1)?.close(3001, 'Acceptance reconnect')
  },
  beginLostReply(requestId: string) {
    pending = outcome(() => db.mutate('increment', null, { requestId }))
  },
  discardReply() {
    discardReplies = true
  },
  recoverReply() {
    discardReplies = false
    sockets.at(-1)?.close(3001, 'Acceptance reply recovery')
  },
  result() {
    return pending
  },
  async close() {
    for (const stop of subscriptions.values()) stop()
    subscriptions.clear()
    await db?.close()
  },
}

declare global {
  interface Window {
    fixture: typeof fixture
  }
}
window.fixture = fixture
