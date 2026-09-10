import type { ReactNode } from 'react'

import { AppLogo } from '@s4wave/web/images/AppLogo.js'

import { LoadingInline } from './LoadingInline.js'
import type { LoadingView } from './types.js'

interface LoadingWorkspaceProps {
  view: Pick<LoadingView, 'title' | 'detail'>
  topLeftSlot?: ReactNode
  children?: ReactNode
  footer?: ReactNode
}

// LoadingWorkspace keeps embedded startup and mount progress inside the workbench.
export function LoadingWorkspace({
  view,
  topLeftSlot,
  children,
  footer,
}: LoadingWorkspaceProps) {
  return (
    <div
      role="status"
      aria-live="polite"
      className="bg-background relative flex h-full min-h-0 flex-1 items-center justify-center p-6"
    >
      {topLeftSlot}
      <div className="flex max-w-full flex-col items-center gap-5 text-center">
        <div
          className="relative flex size-20 items-center justify-center"
          aria-hidden="true"
        >
          <div className="border-brand/15 border-t-brand/60 absolute inset-0 rounded-full border motion-safe:animate-spin motion-safe:[animation-duration:8s]" />
          <AppLogo
            className="w-14 motion-safe:animate-pulse motion-safe:[animation-duration:3s]"
            alt=""
          />
        </div>
        <div className="space-y-2">
          <div className="text-brand/80 text-[10px] font-medium tracking-[0.24em] uppercase">
            Spacewave
          </div>
          <h2 className="text-foreground text-base font-medium">
            {view.title}
          </h2>
          {view.detail && (
            <LoadingInline label={view.detail} className="max-w-full" />
          )}
        </div>
        {children}
        {footer}
      </div>
    </div>
  )
}
