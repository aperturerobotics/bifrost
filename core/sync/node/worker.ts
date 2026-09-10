import { isMainThread, parentPort, workerData } from 'node:worker_threads'
import { createHandler, createMux, Server } from 'starpc'
import { ValueOf } from '@goscript/syscall/js/index.js'

import { SqliteNodeServer } from '../../../db/sql/sqlite-node/server.js'
import { SqliteBridgeDefinition } from '../../../db/sql/sqlite-wasm/rpc/sqlite-bridge_srpc.pb.js'
import { messagePortPacketStream } from '../../../bldr/web/entrypoint/browser/message-port-packet-stream.js'
import { Open } from '@goscript/github.com/s4wave/spacewave/core/sync/node/runtime.gs.js'
import { acquireDirectoryLock } from '../../../db/sql/sqlite-node/lock.js'
import { RuntimeInit, RuntimeMessage, RuntimeMessage_Kind } from './node.pb.js'

// openEngine connects the compiled engine to SQLite within this Worker.
export async function openEngine(directory: string) {
  const sqlite = new SqliteNodeServer()
  const mux = createMux()
  mux.register(createHandler(SqliteBridgeDefinition, sqlite))
  const server = new Server(mux.lookupMethod)
  const openPort = () => {
    const { port1, port2 } = new MessageChannel()
    server.handlePacketStream(messagePortPacketStream(port1))
    return port2
  }
  try {
    const [runtime, error] = await Open(directory, ValueOf(openPort))
    if (error) throw new Error(await error.Error())
    if (!runtime) throw new Error('Sync engine returned no runtime')
    return { runtime, sqlite }
  } catch (error) {
    sqlite.close()
    throw error
  }
}

// runWorker publishes readiness only after SQLite and the World are available.
async function runWorker(): Promise<void> {
  if (!parentPort) throw new Error('Sync engine requires a Worker parent')
  const parent = parentPort
  const init = RuntimeInit.fromBinary(workerData)
  const lock = acquireDirectoryLock(init.directory ?? '')
  let opened: Awaited<ReturnType<typeof openEngine>> | undefined
  try {
    opened = await openEngine(lock.directory)
    const { runtime, sqlite } = opened
    let closing: Promise<void> | undefined
    parent.on(
      'message',
      (frame: { message: Uint8Array; port?: MessagePort }) => {
        const message = RuntimeMessage.fromBinary(frame.message)
        if (message.kind === RuntimeMessage_Kind.OPEN_STREAM && frame.port) {
          if (closing) {
            frame.port.close()
            return
          }
          void Promise.resolve(runtime.Accept(ValueOf(frame.port)))
            .then((error) => {
              if (error) frame.port?.close()
            })
            .catch(() => frame.port?.close())
        } else if (message.kind === RuntimeMessage_Kind.CLOSE) {
          closing ??= (async () => {
            let failure = ''
            try {
              const error = await runtime.Close()
              if (error) throw new Error(await error.Error())
            } catch (error) {
              failure = error instanceof Error ? error.message : String(error)
            } finally {
              try {
                sqlite.close()
              } catch (error) {
                failure ||=
                  error instanceof Error ? error.message : String(error)
              } finally {
                try {
                  lock.close()
                } catch (error) {
                  failure ||=
                    error instanceof Error ? error.message : String(error)
                }
              }
            }
            parent.postMessage(
              RuntimeMessage.toBinary({
                kind: RuntimeMessage_Kind.CLOSED,
                error: failure,
              }),
            )
            parent.close()
          })()
        } else {
          frame.port?.close()
          parent.postMessage(
            RuntimeMessage.toBinary({
              kind: RuntimeMessage_Kind.FAILED,
              error: 'Unsupported sync engine lifecycle request',
            }),
          )
        }
      },
    )
    parent.postMessage(
      RuntimeMessage.toBinary({ kind: RuntimeMessage_Kind.READY }),
    )
  } catch (error) {
    try {
      opened?.sqlite.close()
    } finally {
      lock.close()
    }
    throw error
  }
}

if (!isMainThread) {
  void runWorker().catch((error: unknown) => {
    parentPort?.postMessage(
      RuntimeMessage.toBinary({
        kind: RuntimeMessage_Kind.FAILED,
        error: error instanceof Error ? error.message : String(error),
      }),
    )
    parentPort?.close()
  })
}
