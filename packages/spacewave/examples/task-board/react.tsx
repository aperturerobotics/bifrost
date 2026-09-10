import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { createRoot } from 'react-dom/client'
import { connect, SyncError, type ConnectionState } from 'spacewave'
import { createSyncContext } from 'spacewave/react'
import { schema } from './schema.ts'

const { SyncProvider, useDatabase, useCollection } = createSyncContext(schema)

// UncertainWrite keeps the exact original write closure and requestId so Retry
// resubmits the identical operation; no offline queue is kept.
interface UncertainWrite {
  requestId: string
  message: string
  retry: () => void
}

const connectionLabels: Record<ConnectionState['status'], string> = {
  ready: 'Connected',
  reconnecting: 'Reconnecting…',
  closed: 'Disconnected',
}

// TaskBoard renders the live task list and issues every write against the
// supplied db. Record state comes only from useCollection; a write never
// inserts a row before the server accepts it.
function TaskBoard() {
  const db = useDatabase()
  const todos = useCollection('todos')
  const [connection, setConnection] = useState<ConnectionState>(
    db.connection.current,
  )
  const [title, setTitle] = useState('')
  const [pending, setPending] = useState(false)
  const [uncertain, setUncertain] = useState<UncertainWrite | null>(null)
  const [writeError, setWriteError] = useState<string | null>(null)

  useEffect(() => db.connection.subscribe(setConnection), [db])

  // runWrite runs one write at a time and reports UNCERTAIN with its retry.
  const runWrite = useCallback(
    async (run: () => Promise<unknown>, requestId: string): Promise<void> => {
      setPending(true)
      setUncertain(null)
      setWriteError(null)
      try {
        await run()
      } catch (error) {
        if (error instanceof SyncError && error.code === 'UNCERTAIN') {
          setUncertain({
            requestId,
            message: error.message,
            retry: () => {
              void runWrite(run, requestId)
            },
          })
        } else {
          setWriteError(error instanceof Error ? error.message : String(error))
        }
      } finally {
        setPending(false)
      }
    },
    [],
  )

  const addTask = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const trimmed = title.trim()
    if (!trimmed || pending || connection.status !== 'ready') return
    const id = crypto.randomUUID()
    const requestId = crypto.randomUUID()
    void runWrite(async () => {
      await db
        .collection('todos')
        .put(id, { title: trimmed, done: false }, { requestId })
      setTitle('')
    }, requestId)
  }

  const completeTask = (id: string) => {
    const requestId = crypto.randomUUID()
    void runWrite(
      () => db.mutate('completeTodo', { id }, { requestId }),
      requestId,
    )
  }

  const todosStatus =
    todos.status === 'loading'
      ? 'Loading tasks…'
      : todos.status === 'current'
        ? `Live · ${todos.data.length} ${
            todos.data.length === 1 ? 'task' : 'tasks'
          }`
        : 'Showing the last saved view while reconnecting…'

  return (
    <>
      <h1>Task board</h1>
      <p>
        Two tabs share one server: the <a href="/">Vanilla</a> example and this{' '}
        <a href="/react">React</a> example.
      </p>
      <p className="status">
        {`Connection: ${connectionLabels[connection.status]}`}
        {connection.error ? `: ${connection.error.message}` : ''}
      </p>
      {todos.status === 'error' && todos.error ? (
        <p className="notice" role="alert">
          {todos.error.message}
        </p>
      ) : (
        <p className="status">{todosStatus}</p>
      )}
      <form onSubmit={addTask}>
        <label htmlFor="new-task">
          Add a task
          <input
            id="new-task"
            value={title}
            onChange={(event) => setTitle(event.target.value)}
            disabled={pending}
            required
            maxLength={120}
          />
        </label>
        <button
          className="button"
          type="submit"
          disabled={pending || connection.status !== 'ready' || !title.trim()}
        >
          Add task
        </button>
      </form>
      {writeError ? <p className="notice">{writeError}</p> : null}
      {uncertain ? (
        <p className="notice">
          {uncertain.message} The write may not have been accepted (request{' '}
          {uncertain.requestId}).{' '}
          <button
            className="button"
            type="button"
            disabled={pending || connection.status !== 'ready'}
            onClick={uncertain.retry}
          >
            Retry
          </button>
        </p>
      ) : null}
      <ul className="tasks">
        {todos.data.map((entry) => (
          <li className="task" key={entry.key} data-done={entry.value.done}>
            <span>{entry.value.title}</span>{' '}
            <button
              className="button"
              type="button"
              disabled={
                pending || connection.status !== 'ready' || entry.value.done
              }
              onClick={() => completeTask(entry.key)}
            >
              {entry.value.done ? 'Completed' : 'Complete'}
            </button>
          </li>
        ))}
      </ul>
    </>
  )
}

// start connects, mounts the board into #app, and registers pagehide teardown.
async function start(): Promise<void> {
  const container = document.getElementById('app')
  if (!container) throw new Error('Missing #app mount point')
  const url = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${
    location.host
  }/sync`
  const db = await connect({
    url,
    schema,
    getAccessToken: () => 'local-demo',
  })
  const root = createRoot(container)
  root.render(
    <SyncProvider db={db}>
      <TaskBoard />
    </SyncProvider>,
  )
  window.addEventListener(
    'pagehide',
    () => {
      root.unmount()
      void db.close()
    },
    { once: true },
  )
}

start().catch((error: unknown) => {
  const container = document.getElementById('app') ?? document.body
  const notice = document.createElement('p')
  notice.className = 'notice'
  notice.textContent = `Task board failed to start: ${
    error instanceof Error ? error.message : String(error)
  }`
  container.replaceChildren(notice)
})
