import { pushable, type Pushable } from 'it-pushable'
import { useCallback, useMemo, useRef, useState } from 'react'
import { LuAppWindow } from 'react-icons/lu'

import { WasiVmTypeID, type WasiVmHandle } from '@s4wave/sdk/vm/wasivm.js'
import {
  WasiVmWindowPlace,
  type WasiVmWindow,
  type WasiVmWindowInput,
} from '@s4wave/sdk/vm/wasivm.pb.js'
import { createObjectLinkAddTabRequest } from '@s4wave/web/object/ObjectLink.js'
import { useTabContext } from '@s4wave/web/object/TabContext.js'
import { useNavigate, usePath } from '@s4wave/web/router/router.js'
import { Button } from '@s4wave/web/ui/button.js'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from '@s4wave/web/ui/DropdownMenu.js'

import {
  WasiVmDisplay,
  type WasiVmDisplayConnector,
  type WasiVmDisplayViews,
  type WasiVmPaneSize,
} from './WasiVmDisplay.js'

// WasiVmWindowRoute names one window of one run of a VM.
export interface WasiVmWindowRoute {
  run: bigint
  id: number
}

// windowPathPattern matches a window's viewer path, window/<run>/<id>.
const windowPathPattern = /^\/?window\/(\d+)\/(\d+)\/?$/

// parseWindowPath reads the window a viewer path names, or null when it
// names the VM.
export function parseWindowPath(path: string): WasiVmWindowRoute | null {
  const m = windowPathPattern.exec(path)
  return m ? { run: BigInt(m[1]), id: Number(m[2]) } : null
}

// formatWindowPath returns the viewer path of a window.
export function formatWindowPath(route: WasiVmWindowRoute): string {
  return `window/${route.run}/${route.id}`
}

// windowLabel names a window by its title, else its app id, else its id.
export function windowLabel(w: WasiVmWindow): string {
  return w.title || w.appId || `Window ${w.id ?? 0}`
}

// WasiVmWindowView is the window a VM viewer's path names, with the actions
// that move between the VM's view and its windows' tabs.
export interface WasiVmWindowView {
  // route is the window the path names, or null for the VM's view.
  route: WasiVmWindowRoute | null
  // shownWindow is the named window while this run lists it.
  shownWindow?: WasiVmWindow
  // showDesktop returns the viewer to the VM's view.
  showDesktop: () => void
  // openWindow opens a window of this run in a tab of its own, or in this
  // viewer outside a tab host.
  openWindow: (id: number) => void
}

// useWasiVmWindowRoute reads the window the viewer path names among the
// windows of run runGeneration of the VM objectKey.
export function useWasiVmWindowRoute(
  objectKey: string,
  runGeneration: bigint,
  windows: WasiVmWindow[],
): WasiVmWindowView {
  const route = parseWindowPath(usePath())
  const navigate = useNavigate()
  const tabContext = useTabContext()
  const shownWindow =
    route?.run === runGeneration
      ? windows.find((w) => w.id === route.id)
      : undefined

  const showDesktop = useCallback(() => navigate({ path: '/' }), [navigate])
  const openWindow = (id: number) => {
    const path = formatWindowPath({ run: runGeneration, id })
    if (!tabContext) {
      navigate({ path: '/' + path })
      return
    }
    const w = windows.find((w) => w.id === id)
    void tabContext
      .addTab(
        createObjectLinkAddTabRequest({
          objectKey,
          objectType: WasiVmTypeID,
          label: w ? windowLabel(w) : undefined,
          path,
          currentTabId: tabContext.tabId,
        }),
      )
      .catch((err) => console.warn('WasiVmWindow: failed to add tab:', err))
  }
  return { route, shownWindow, showDesktop, openWindow }
}

// WasiVmWindowsMenuProps configures the windows menu of a running VM.
interface WasiVmWindowsMenuProps {
  windows: WasiVmWindow[]
  onOpen: (id: number) => void
}

// WasiVmWindowsMenu lists the windows of a VM's Wayland display, each opening
// in a tab of its own. Its label collapses to the icon in a narrow header.
export function WasiVmWindowsMenu({ windows, onOpen }: WasiVmWindowsMenuProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="toolbarSm" aria-label="Windows">
          <LuAppWindow />
          <span className="hidden @md:inline">Windows</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" variant="compact" className="min-w-45">
        <DropdownMenuLabel>Open in tab</DropdownMenuLabel>
        {windows.length === 0 ? (
          <DropdownMenuItem disabled>No windows are open</DropdownMenuItem>
        ) : (
          windows.map((w) => (
            <DropdownMenuItem key={w.id} onClick={() => onOpen(w.id ?? 0)}>
              <span className="min-w-0 flex-1 truncate">{windowLabel(w)}</span>
              {w.place === WasiVmWindowPlace.WasiVmWindowPlace_VIEWER && (
                <span className="text-foreground-alt/60 text-xs">In a tab</span>
              )}
            </DropdownMenuItem>
          ))
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

// WasiVmWindowPaneProps configures the pane showing one window.
interface WasiVmWindowPaneProps {
  handle: WasiVmHandle | undefined
  id: number
  onShowDesktop: () => void
}

// WasiVmWindowPane shows one window of a running VM's Wayland display at the
// pane's size. It opens the window once the pane is measured and resizes it
// as the pane changes; leaving returns the window to the desktop.
export function WasiVmWindowPane({
  handle,
  id,
  onShowDesktop,
}: WasiVmWindowPaneProps) {
  // sizeRef is the pane's last size; windowInput is the open stream's input.
  const sizeRef = useRef<WasiVmPaneSize | null>(null)
  const windowInput = useRef<Pushable<WasiVmWindowInput> | null>(null)
  const [measured, setMeasured] = useState(false)

  // Open the window at the pane's size, passing on the display's input.
  const connectWindow = useMemo<WasiVmDisplayConnector | undefined>(() => {
    if (!handle || !measured) return undefined
    return (input, signal) => {
      const win = pushable<WasiVmWindowInput>({ objectMode: true })
      win.push({
        body: {
          case: 'open',
          value: { id, size: sizeRef.current ?? undefined },
        },
      })
      windowInput.current = win
      void (async () => {
        for await (const ev of input) {
          win.push({ body: { case: 'input', value: ev } })
        }
        win.end()
        if (windowInput.current === win) windowInput.current = null
      })()
      return handle.connectWindow(win, signal)
    }
  }, [handle, measured, id])

  // Resize the open window with the pane.
  const onPaneSize = useCallback((size: WasiVmPaneSize) => {
    sizeRef.current = size
    setMeasured(true)
    windowInput.current?.push({ body: { case: 'resize', value: size } })
  }, [])

  const views = useMemo<WasiVmDisplayViews>(
    () => ({
      waiting: {
        state: 'active',
        title: 'Opening the window',
        detail: 'It appears here once the VM draws it at this size.',
      },
      closed: () => ({
        state: 'error',
        title: 'The window closed',
        detail: 'Its app closed it, or the VM stopped its display.',
        onRetry: onShowDesktop,
        retryLabel: 'Show desktop',
      }),
      failed: (detail, retry) => ({
        state: 'error',
        title: 'The window disconnected',
        detail,
        onRetry: retry,
        onCancel: onShowDesktop,
        cancelLabel: 'Show desktop',
      }),
    }),
    [onShowDesktop],
  )

  return (
    <WasiVmDisplay
      connectDisplay={connectWindow}
      views={views}
      onPaneSize={onPaneSize}
    />
  )
}
