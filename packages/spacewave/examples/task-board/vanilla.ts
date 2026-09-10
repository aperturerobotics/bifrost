import { connect, SyncError } from 'spacewave'
import { schema } from './schema.ts'

const app = document.querySelector<HTMLElement>('#app')!
app.innerHTML =
  '<nav><a href="/">Vanilla</a><a href="/react">React</a></nav><h1>Task board</h1><p>Open two tabs. Add or complete a task and watch both boards update.</p><div class="status" role="status">Connecting…</div><form><label>Add a task<input name="title" required maxlength="120" autocomplete="off"></label><button class="button">Add task</button></form><div class="notice" role="alert" hidden></div><p id="freshness" role="status"></p><ul class="tasks"></ul>'
const form = app.querySelector<HTMLFormElement>('form')!
const input = app.querySelector<HTMLInputElement>('input')!
const status = app.querySelector<HTMLElement>('.status')!
const notice = app.querySelector<HTMLElement>('.notice')!
const freshness = app.querySelector<HTMLElement>('#freshness')!
const list = app.querySelector<HTMLElement>('.tasks')!
let pending = false
let ready = false
const updateButtons = () => {
  for (const button of app.querySelectorAll<HTMLButtonElement>('button')) {
    button.disabled = pending || !ready || button.dataset.done === 'true'
  }
}

async function start(): Promise<void> {
  const url = new URL('/sync', location.href)
  url.protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const db = await connect({
    url: url.href,
    schema,
    getAccessToken: () => 'local-demo',
  })
  const todos = db.collection('todos')
  const stopConnection = db.connection.subscribe((connection) => {
    ready = connection.status === 'ready'
    status.textContent =
      connection.status === 'ready'
        ? 'Connected'
        : connection.status === 'reconnecting'
          ? 'Reconnecting…'
          : 'Connection closed'
    updateButtons()
  })
  const runWrite = async (write: () => Promise<unknown>) => {
    pending = true
    notice.hidden = true
    updateButtons()
    try {
      await write()
    } catch (error) {
      notice.hidden = false
      notice.replaceChildren(
        document.createTextNode(
          error instanceof Error
            ? error.message
            : 'The change could not complete',
        ),
      )
      if (error instanceof SyncError && error.code === 'UNCERTAIN') {
        notice.append(document.createTextNode(` Request: ${error.requestId}. `))
        const retry = document.createElement('button')
        retry.textContent = 'Retry'
        retry.onclick = () => {
          void runWrite(write)
        }
        notice.append(retry)
      }
    } finally {
      pending = false
      updateButtons()
    }
  }
  const stop = todos.subscribe(
    {},
    {
      next: (state) => {
        freshness.textContent =
          state.status === 'loading'
            ? 'Loading tasks…'
            : state.status === 'stale'
              ? 'Showing the last saved view while reconnecting…'
              : state.status === 'error'
                ? (state.error?.message ?? 'Tasks could not load')
                : state.data.length
                  ? 'All changes saved'
                  : 'No tasks yet'
        list.replaceChildren(
          ...state.data.map(({ key, value }) => {
            const row = document.createElement('li')
            row.className = 'task'
            row.dataset.done = String(value.done)
            const title = document.createElement('span')
            title.textContent = value.title
            const complete = document.createElement('button')
            complete.textContent = value.done ? 'Completed' : 'Complete'
            complete.dataset.done = String(value.done)
            complete.onclick = () => {
              const requestId = crypto.randomUUID()
              void runWrite(() =>
                db.mutate('completeTodo', { id: key }, { requestId }),
              )
            }
            row.append(title, complete)
            return row
          }),
        )
        updateButtons()
      },
    },
  )
  form.onsubmit = (event) => {
    event.preventDefault()
    const title = input.value.trim()
    if (!title || pending || !ready) return
    const key = crypto.randomUUID()
    const requestId = crypto.randomUUID()
    void runWrite(async () => {
      await todos.put(key, { title, done: false }, { requestId })
      input.value = ''
    })
  }
  window.addEventListener(
    'pagehide',
    () => {
      stop()
      stopConnection()
      void db.close()
    },
    { once: true },
  )
}
updateButtons()
void start().catch((error: unknown) => {
  status.textContent = 'Connection failed'
  notice.hidden = false
  notice.textContent =
    error instanceof Error ? error.message : 'The board could not connect'
})
