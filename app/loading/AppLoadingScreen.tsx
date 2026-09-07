import { useEffect, type ReactNode } from 'react'

import { useBrowserStartupProjection } from '@s4wave/app/loading/status/browser-startup.js'
import { useBrowserBootWatchdog } from '@s4wave/app/loading/status/browser-boot-watchdog.js'
import { markBrowserStartupBoundary } from '@s4wave/app/prerender/boot-status.js'
import { LoadingScreen } from '@s4wave/web/ui/loading/LoadingScreen.js'

import { BootLoadingCriticalStyle } from './boot-loading-critical.js'
import { BrowserStartupPhaseRail } from './BrowserStartupPhaseRail.js'
import { useBootDownloads } from './status/browser-downloads.js'

// AppLoadingScreen presents startup activity before the application stylesheet
// arrives. Readiness and download evidence stay in the startup projection.
export function AppLoadingScreen() {
  const startup = useBrowserStartupProjection()
  useBrowserBootWatchdog()
  const view = startup.view
  const downloads = useBootDownloads()
  const failedDownload = downloads.find(
    (download) => download.state === 'error',
  )
  const transfer = downloads.find(
    (download) => download.state === 'active' && (download.total ?? 0) > 0,
  )
  const failed = view.state === 'error' || failedDownload !== undefined
  const error =
    view.error ??
    failedDownload?.error ??
    (failedDownload
      ? `${failedDownload.label} could not be loaded.`
      : undefined)

  return (
    <BrowserStartupRevealProbe>
      <BootLoadingCriticalStyle />
      <LoadingScreen
        view={{
          state: failed ? 'error' : 'loading',
          title: failed ? 'Unable to open Spacewave' : view.title,
          detail: view.detail ?? startup.phase.label,
          error,
          progress: transfer?.total
            ? transfer.loaded / transfer.total
            : undefined,
          onRetry: failed ? retryBrowserStartup : undefined,
          onCancel: leaveBrowserStartup,
          cancelLabel: failed ? 'Back' : 'Back to home',
        }}
      >
        <BrowserStartupPhaseRail phases={startup.phases} />
      </LoadingScreen>
    </BrowserStartupRevealProbe>
  )
}

function retryBrowserStartup() {
  markBrowserStartupBoundary('webview.loading-surface-retry', {
    source: 'app',
  })
  window.location.reload()
}

function leaveBrowserStartup() {
  markBrowserStartupBoundary('webview.loading-surface-back', {
    source: 'app',
  })
  if (window.history.length > 1) {
    window.history.back()
    return
  }
  localStorage.removeItem('spacewave-has-session')
  window.location.assign('/')
}

function BrowserStartupRevealProbe({ children }: { children: ReactNode }) {
  useEffect(() => {
    markBrowserStartupBoundary('webview.loading-surface-mounted', {
      source: 'app',
    })
    return () => {
      markBrowserStartupBoundary('webview.loading-surface-revealed', {
        source: 'app',
      })
    }
  }, [])

  return children
}
