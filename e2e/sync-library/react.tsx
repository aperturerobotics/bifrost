import { createRoot, type Root } from 'react-dom/client'
import { connect, type Database } from 'spacewave'
import { createSyncContext } from 'spacewave/react'
import { schema } from './schema.js'

const { SyncProvider, useCollection } = createSyncContext(schema)
let db: Database<typeof schema>
let root: Root | undefined
let active = 0
let opened = 0
let prefix = ''

function View() {
  const state = useCollection('todos', { prefix })
  return (
    <p id="result" data-status={state.status}>
      {state.data.map((record) => record.key).join(',')}
    </p>
  )
}

declare global {
  interface Window {
    reactFixture: {
      start(url: string): Promise<void>
      mount(): void
      unmount(): void
      prefix(prefix: string): void
      counts(): { active: number; opened: number }
      close(): Promise<void>
    }
  }
}

window.reactFixture = {
  async start(url) {
    db = await connect({ url, schema, getAccessToken: () => 'alice' })
    const collection = db.collection.bind(db)
    db.collection = ((name) => {
      const handle = collection(name)
      const watch = handle.watch.bind(handle)
      handle.watch = async function* (...args) {
        active++
        opened++
        try {
          yield* watch(...args)
        } finally {
          active--
        }
      }
      return handle
    }) as typeof db.collection
  },
  mount() {
    root ??= createRoot(document.querySelector('#app')!)
    root.render(
      <SyncProvider db={db}>
        <View />
      </SyncProvider>,
    )
  },
  unmount() {
    root?.unmount()
    root = undefined
  },
  prefix(value) {
    prefix = value
    this.mount()
  },
  counts() {
    return { active, opened }
  },
  async close() {
    this.unmount()
    await db.close()
  },
}
