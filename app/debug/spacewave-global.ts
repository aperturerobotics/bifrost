import { pluginPathPrefix } from '@s4wave/app/urls.js'
import { resolvePath, To } from '@s4wave/web/router/router.js'
import { getDebugContext } from '@s4wave/sdk/debug/context.js'

// SpacewaveDebug provides debug commands on window.spacewave.
interface SpacewaveDebug {
  // app returns one mounted app's normal SDK and scenario controls.
  app: (id: string) => ReturnType<typeof getDebugContext>
  // navigate resolves a path (relative or absolute) against the current
  // hash path and navigates to it. Supports ../../, ./foo/bar, /absolute.
  navigate: (to: string | To) => void
  // downloadTrace captures and downloads a Go runtime trace.
  downloadTrace: (seconds?: number) => void
  // path returns the current hash path.
  readonly path: string
}

function getHashPath(): string {
  const hash = window.location.hash
  return hash.startsWith('#') ? hash.slice(1) : hash || '/'
}

function triggerDownload(path: string, filename: string): void {
  const a = document.createElement('a')
  a.href = pluginPathPrefix + path
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

const spacewave: SpacewaveDebug = {
  app: getDebugContext,
  navigate(to: string | To) {
    const toObj = typeof to === 'string' ? { path: to } : to
    const resolved = resolvePath(getHashPath(), toObj)
    if (toObj.replace) {
      window.location.replace('#' + resolved)
    } else {
      window.location.hash = resolved
    }
  },
  downloadTrace(seconds = 30) {
    const duration = Number.isFinite(seconds) && seconds > 0 ? seconds : 30
    triggerDownload(`/debugz/trace?seconds=${duration}`, 'trace.out')
  },
  get path() {
    return getHashPath()
  },
}

;(window as unknown as Record<string, unknown>)['spacewave'] = spacewave
