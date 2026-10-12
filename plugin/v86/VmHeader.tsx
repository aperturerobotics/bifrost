import {
  LuKeyboard,
  LuMaximize,
  LuMonitor,
  LuSquare,
  LuTerminal,
  LuTv,
} from 'react-icons/lu'

import { Button } from '@s4wave/web/ui/button.js'
import { cn } from '@s4wave/web/style/utils.js'
import type { VmState } from '@s4wave/sdk/vm/v86.pb.js'

import { VmStateBadge } from './VmStateBadge.js'

// VmView is the panel the running view shows.
export type VmView = 'screen' | 'console'

// VmHeaderControls are the controls of a VM that is running or about to.
export interface VmHeaderControls {
  view: VmView
  onViewChange: (view: VmView) => void
  // stopDisabled is set once a stop is requested.
  stopDisabled: boolean
  onStop: () => void
  onCtrlAltDel: () => void
  onFullscreen: () => void
}

// VmHeaderProps configures the header strip of a VM.
export interface VmHeaderProps {
  name: string
  state: VmState
  // controls is set while the VM runs; a stopped VM shows only its identity.
  controls?: VmHeaderControls
}

// viewTabs lists the panels of the running view.
const viewTabs = [
  { view: 'screen', label: 'Screen', icon: LuMonitor },
  { view: 'console', label: 'Console', icon: LuTerminal },
] as const

// VmHeader is the slim strip above a VM: its name and state and, while it
// runs, the panel tabs and machine actions. Labels collapse to icons in a
// narrow panel.
export function VmHeader({ name, state, controls }: VmHeaderProps) {
  return (
    <div className="border-foreground/8 @container flex h-9 shrink-0 items-center gap-2 border-b px-3">
      <div className="flex min-w-0 flex-1 items-center gap-2 select-none">
        <LuTv className="text-foreground-alt size-4 shrink-0" />
        <span className="text-foreground truncate text-sm font-medium tracking-tight">
          {name}
        </span>
        <VmStateBadge state={state} />
      </div>
      {controls && (
        <>
          <div
            role="group"
            aria-label="VM panel"
            className="bg-foreground/5 flex shrink-0 rounded-md p-0.5"
          >
            {viewTabs.map(({ view, label, icon: Icon }) => (
              <button
                key={view}
                type="button"
                aria-pressed={controls.view === view}
                aria-label={label}
                onClick={() => controls.onViewChange(view)}
                className={cn(
                  'flex h-6 items-center gap-1.5 rounded px-2 text-xs transition-colors',
                  controls.view === view
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-foreground-alt hover:text-foreground',
                )}
              >
                <Icon className="size-3.5" />
                <span className="hidden @lg:inline">{label}</span>
              </button>
            ))}
          </div>
          <div className="flex shrink-0 items-center gap-1">
            <Button
              variant="ghost"
              size="toolbarSm"
              aria-label="Send Ctrl+Alt+Delete"
              title="Send Ctrl+Alt+Delete"
              onClick={controls.onCtrlAltDel}
            >
              <LuKeyboard />
              <span className="hidden @2xl:inline">Ctrl+Alt+Del</span>
            </Button>
            <Button
              variant="ghost"
              size="toolbarSm"
              aria-label="Fullscreen"
              title="Fullscreen"
              onClick={controls.onFullscreen}
            >
              <LuMaximize />
            </Button>
            <Button
              variant="destructiveOutline"
              size="toolbarSm"
              aria-label="Stop"
              disabled={controls.stopDisabled}
              onClick={controls.onStop}
            >
              <LuSquare />
              <span className="hidden @md:inline">Stop</span>
            </Button>
          </div>
        </>
      )}
    </div>
  )
}
