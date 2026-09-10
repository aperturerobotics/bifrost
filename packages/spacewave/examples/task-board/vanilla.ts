import { connect, SyncError } from 'spacewave'

import { schema } from './schema.ts'

// Keep the fixed page scaffold separate from task values rendered with textContent.
const app = document.querySelector<HTMLElement>('#app')!
app.innerHTML = `
  <nav>
    <a href="/">Vanilla</a>
    <a href="/react">React</a>
  </nav>

  <h1>Task board</h1>
  <p>Open two tabs. Add or complete a task and watch both boards update.</p>
  <div class="status" role="status">Connecting…</div>

  <form>
    <label>
      Add a task
      <input name="title" required maxlength="120" autocomplete="off">
    </label>
    <button class="button">Add task</button>
  </form>

  <div class="notice" role="alert" hidden></div>
  <p id="freshness" role="status"></p>
  <ul class="tasks"></ul>
`

// Bind the controls used by connection, write, and subscription callbacks.
const form = app.querySelector<HTMLFormElement>('form')!
const input = app.querySelector<HTMLInputElement>('input')!
const status = app.querySelector<HTMLElement>('.status')!
const notice = app.querySelector<HTMLElement>('.notice')!
const freshness = app.querySelector<HTMLElement>('#freshness')!
const list = app.querySelector<HTMLElement>('.tasks')!

// Pending form work stays separate from the accepted collection snapshot.
let pending = false
let ready = false

/** updateButtons permits edits only while connected and without a pending write. */
function updateButtons(): void {
  for (const button of app.querySelectorAll<HTMLButtonElement>('button')) {
    button.disabled = pending || !ready || button.dataset.done === 'true'
  }
}

/** runWrite retains the original write closure so Retry preserves its input and ID. */
async function runWrite(write: () => Promise<unknown>): Promise<void> {
  // Suspend edits while the server decides this write's result.
  pending = true
  notice.hidden = true
  updateButtons()

  // Show failures without changing the accepted task rows.
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

    // Retrying an uncertain write reuses its original input and request ID.
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

/** start connects the board and releases its subscriptions when the page leaves. */
async function start(): Promise<void> {
  // Open one connection for the page and use its typed task collection.
  const url = new URL('/sync', location.href)
  url.protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const db = await connect({
    url: url.href,
    schema,
    getAccessToken: () => 'local-demo',
  })
  const todos = db.collection('todos')

  // Keep connection feedback and edit availability in sync.
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

  // Replace the rendered list with each complete server snapshot.
  const stop = todos.subscribe(
    {},
    {
      next: (state) => {
        // Describe freshness separately from the retained task values.
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

        // Render task titles as text and wire each row's completion action.
        list.replaceChildren(
          ...state.data.map(({ key, value }) => {
            // Build the row from the accepted task value.
            const row = document.createElement('li')
            row.className = 'task'
            row.dataset.done = String(value.done)
            const title = document.createElement('span')
            title.textContent = value.title

            // Keep a completion request's ID stable if the user retries it.
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

  // Capture a new task's input once so recovery can repeat the exact write.
  form.onsubmit = (event) => {
    // Accept a nonempty title only when edits are available.
    event.preventDefault()
    const title = input.value.trim()
    if (!title || pending || !ready) return

    // Clear the input after acceptance, preserving the IDs inside the closure.
    const key = crypto.randomUUID()
    const requestId = crypto.randomUUID()
    void runWrite(async () => {
      await todos.put(key, { title, done: false }, { requestId })
      input.value = ''
    })
  }

  // Release both subscriptions before closing the page's database.
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

// Disable edits until admission succeeds and make startup failures visible.
updateButtons()
void start().catch((error: unknown) => {
  status.textContent = 'Connection failed'
  notice.hidden = false
  notice.textContent =
    error instanceof Error ? error.message : 'The board could not connect'
})
