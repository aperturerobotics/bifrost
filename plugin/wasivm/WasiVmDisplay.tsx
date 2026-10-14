import { pushable, type Pushable } from 'it-pushable'
import { useEffect, useReducer, useRef, useState } from 'react'
import type { MessageStream } from 'starpc'

import type {
  WasiVmDisplayFrame,
  WasiVmDisplayInput,
} from '@s4wave/sdk/vm/wasivm.pb.js'
import { cn } from '@s4wave/web/style/utils.js'
import { LoadingCard } from '@s4wave/web/ui/loading/LoadingCard.js'

import { keysymOf, xButtons } from './display-input.js'

// WasiVmDisplayConnector opens the display stream of a running VM.
export type WasiVmDisplayConnector = (
  input: MessageStream<WasiVmDisplayInput>,
  signal: AbortSignal,
) => MessageStream<WasiVmDisplayFrame>

// WasiVmDisplayProps configures the display of a VM.
export interface WasiVmDisplayProps {
  connectDisplay?: WasiVmDisplayConnector
}

// DisplayStatus is the state of the display stream: waiting for the first
// frame, showing the screen, or ended with a reason.
type DisplayStatus =
  | { kind: 'waiting' }
  | { kind: 'ready' }
  | { kind: 'ended'; detail: string }

// wheelUp and wheelDown are the masks of X buttons 4 and 5, which a wheel
// step presses and releases to scroll up and down.
const wheelUp = 1 << 3
const wheelDown = 1 << 4

// WasiVmDisplay shows X display 0 of a running VM on a canvas scaled to fit
// the pane. Pointer, wheel and key input on the canvas drive the desktop.
export function WasiVmDisplay({ connectDisplay }: WasiVmDisplayProps) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null)
  const inputRef = useRef<Pushable<WasiVmDisplayInput> | null>(null)
  // heldKeys maps each held key's code to the keysym its press sent, so the
  // release sends the same keysym after a modifier changes.
  const heldKeys = useRef(new Map<string, number>())
  const [status, setStatus] = useState<DisplayStatus>({ kind: 'waiting' })
  const [attempt, retry] = useReducer((n: number) => n + 1, 0)

  // Stream the screen onto the canvas.
  useEffect(() => {
    const canvas = canvasRef.current
    const ctx = canvas?.getContext('2d')
    if (!connectDisplay || !canvas || !ctx) return
    const abort = new AbortController()
    const input = pushable<WasiVmDisplayInput>({ objectMode: true })
    inputRef.current = input
    setStatus({ kind: 'waiting' })

    // cancelled is set once the effect is cleaned up, after which no state
    // is set.
    let cancelled = false
    void (async () => {
      let ready = false
      try {
        for await (const frame of connectDisplay(input, abort.signal)) {
          // Size the canvas to the screen, which clears it.
          const screenWidth = frame.screenWidth ?? 0
          const screenHeight = frame.screenHeight ?? 0
          if (canvas.width !== screenWidth || canvas.height !== screenHeight) {
            canvas.width = screenWidth
            canvas.height = screenHeight
          }

          // Draw the region.
          const width = frame.width ?? 0
          const height = frame.height ?? 0
          const rgba = frame.rgba
          if (width && height && rgba?.length === width * height * 4) {
            const pixels = new Uint8ClampedArray(rgba)
            ctx.putImageData(
              new ImageData(pixels, width, height),
              frame.x ?? 0,
              frame.y ?? 0,
            )
          }
          if (!ready && !cancelled) {
            ready = true
            setStatus({ kind: 'ready' })
          }
        }
        if (!cancelled) {
          setStatus({
            kind: 'ended',
            detail: 'The X server closed the display.',
          })
        }
      } catch (err) {
        if (!cancelled) {
          const detail = err instanceof Error ? err.message : String(err)
          setStatus({ kind: 'ended', detail })
        }
      }
    })()

    return () => {
      cancelled = true
      abort.abort()
      input.end()
      inputRef.current = null
    }
  }, [connectDisplay, attempt])

  // send queues one input event for the VM.
  const send = (body: WasiVmDisplayInput['body']) => {
    inputRef.current?.push({ body })
  }

  // sendPointer sends the pointer's position on the screen and the held
  // buttons, with extra X buttons added.
  const sendPointer = (
    e: { clientX: number; clientY: number; buttons: number },
    extra = 0,
  ) => {
    const canvas = canvasRef.current
    if (!canvas) return
    const rect = canvas.getBoundingClientRect()
    if (!rect.width || !rect.height) return
    send({
      case: 'pointer',
      value: {
        x: Math.floor(((e.clientX - rect.left) * canvas.width) / rect.width),
        y: Math.floor(((e.clientY - rect.top) * canvas.height) / rect.height),
        buttons: xButtons(e.buttons) | extra,
      },
    })
  }

  // releasePointer releases every button where an interrupted drag left the
  // pointer.
  const releasePointer = (e: { clientX: number; clientY: number }) => {
    sendPointer({ clientX: e.clientX, clientY: e.clientY, buttons: 0 })
  }

  // releaseKeys releases every held key, as when the canvas loses focus.
  const releaseKeys = () => {
    for (const keysym of heldKeys.current.values()) {
      send({ case: 'key', value: { keysym, down: false } })
    }
    heldKeys.current.clear()
  }

  return (
    <div className="relative flex min-h-0 flex-1 items-center justify-center overflow-hidden bg-black">
      <canvas
        ref={canvasRef}
        tabIndex={0}
        aria-label="VM display"
        className={cn(
          'max-h-full max-w-full touch-none outline-none',
          status.kind !== 'ready' && 'invisible',
        )}
        onContextMenu={(e) => e.preventDefault()}
        onPointerDown={(e) => {
          e.currentTarget.setPointerCapture(e.pointerId)
          e.currentTarget.focus()
          sendPointer(e)
        }}
        onPointerMove={(e) => sendPointer(e)}
        onPointerUp={(e) => sendPointer(e)}
        onPointerCancel={releasePointer}
        onLostPointerCapture={releasePointer}
        onWheel={(e) => {
          if (!e.deltaY) return
          sendPointer(e, e.deltaY < 0 ? wheelUp : wheelDown)
          sendPointer(e)
        }}
        onKeyDown={(e) => {
          const keysym = heldKeys.current.get(e.code) ?? keysymOf(e.key, e.code)
          if (!keysym) return
          e.preventDefault()
          heldKeys.current.set(e.code, keysym)
          send({ case: 'key', value: { keysym, down: true } })
        }}
        onKeyUp={(e) => {
          const keysym = heldKeys.current.get(e.code)
          if (keysym === undefined) return
          e.preventDefault()
          heldKeys.current.delete(e.code)
          send({ case: 'key', value: { keysym, down: false } })
        }}
        onBlur={releaseKeys}
      />
      {status.kind !== 'ready' && (
        <LoadingCard
          className="absolute w-full max-w-sm"
          view={
            status.kind === 'waiting'
              ? {
                  state: 'active',
                  title: 'Waiting for the display',
                  detail:
                    'The desktop appears once the VM starts its X server.',
                }
              : {
                  state: 'error',
                  title: 'The display disconnected',
                  detail: status.detail,
                  onRetry: retry,
                }
          }
        />
      )}
    </div>
  )
}
