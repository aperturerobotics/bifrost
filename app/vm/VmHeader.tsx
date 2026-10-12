import type { ReactNode } from 'react'
import { LuSquare, LuTv } from 'react-icons/lu'

import { Button } from '@s4wave/web/ui/button.js'
import type { VmState } from '@s4wave/sdk/vm/v86.pb.js'

import { VmStateBadge } from './VmStateBadge.js'

// VmHeaderProps configures the header strip of a VM.
export interface VmHeaderProps {
  name: string
  state: VmState
  // children are the controls of a VM that is running or about to; a
  // stopped VM shows only its identity.
  children?: ReactNode
}

// VmHeader is the slim strip above a VM: its name and state, then its
// controls. The strip is a container, so controls can collapse their labels
// in a narrow panel; the leading icon hides there too.
export function VmHeader({ name, state, children }: VmHeaderProps) {
  return (
    <div className="border-foreground/8 @container flex h-9 shrink-0 items-center gap-2 border-b px-3">
      <div className="flex min-w-0 flex-1 items-center gap-2 select-none">
        <LuTv className="text-foreground-alt hidden size-4 shrink-0 @md:block" />
        <span className="text-foreground truncate text-sm font-medium tracking-tight">
          {name}
        </span>
        <VmStateBadge state={state} />
      </div>
      {children}
    </div>
  )
}

// VmStopButtonProps configures the stop action of a VM.
export interface VmStopButtonProps {
  // disabled is set once a stop is requested.
  disabled: boolean
  onStop: () => void
}

// VmStopButton asks a running VM to stop. Its label collapses to the icon in
// a narrow header.
export function VmStopButton({ disabled, onStop }: VmStopButtonProps) {
  return (
    <Button
      variant="destructiveOutline"
      size="toolbarSm"
      aria-label="Stop"
      disabled={disabled}
      onClick={onStop}
    >
      <LuSquare />
      <span className="hidden @md:inline">Stop</span>
    </Button>
  )
}
