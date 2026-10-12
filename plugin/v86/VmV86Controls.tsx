import { LuKeyboard, LuMaximize, LuMonitor, LuTerminal } from 'react-icons/lu'

import { VmStopButton } from '@s4wave/app/vm/VmHeader.js'
import { Button } from '@s4wave/web/ui/button.js'
import { cn } from '@s4wave/web/style/utils.js'

// VmView is the panel the running view shows.
export type VmView = 'screen' | 'console'

// VmV86ControlsProps are the controls of a VM that is running or about to.
export interface VmV86ControlsProps {
  view: VmView
  onViewChange: (view: VmView) => void
  // stopDisabled is set once a stop is requested.
  stopDisabled: boolean
  onStop: () => void
  onCtrlAltDel: () => void
  onFullscreen: () => void
}

// viewTabs lists the panels of the running view.
const viewTabs = [
  { view: 'screen', label: 'Screen', icon: LuMonitor },
  { view: 'console', label: 'Console', icon: LuTerminal },
] as const

// VmV86Controls are the panel tabs and machine actions of a running x86 VM,
// shown in its header. Labels collapse to icons in a narrow panel.
export function VmV86Controls({
  view,
  onViewChange,
  stopDisabled,
  onStop,
  onCtrlAltDel,
  onFullscreen,
}: VmV86ControlsProps) {
  return (
    <>
      <div
        role="group"
        aria-label="VM panel"
        className="bg-foreground/5 flex shrink-0 rounded-md p-0.5"
      >
        {viewTabs.map(({ view: tab, label, icon: Icon }) => (
          <button
            key={tab}
            type="button"
            aria-pressed={view === tab}
            aria-label={label}
            onClick={() => onViewChange(tab)}
            className={cn(
              'flex h-6 items-center gap-1.5 rounded px-2 text-xs transition-colors',
              view === tab
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
          onClick={onCtrlAltDel}
        >
          <LuKeyboard />
          <span className="hidden @2xl:inline">Ctrl+Alt+Del</span>
        </Button>
        <Button
          variant="ghost"
          size="toolbarSm"
          aria-label="Fullscreen"
          title="Fullscreen"
          onClick={onFullscreen}
        >
          <LuMaximize />
        </Button>
        <VmStopButton disabled={stopDisabled} onStop={onStop} />
      </div>
    </>
  )
}
