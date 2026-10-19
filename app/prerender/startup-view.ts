import { isPathnameAppRoute } from '@s4wave/web/router/app-path.js'
import { isStaticRoute } from '@s4wave/web/router/static-routes.js'

// StartupView is what the startup module shows while the Go runtime loads:
// a prerendered static page, the root landing page, or the loading screen.
export type StartupView = 'static' | 'landing' | 'loading'

// selectStartupView picks the startup view for a location. Static pages show
// only on their pathname without a hash, except the root, whose template is
// the landing page for new visitors and the loading screen for returning
// visitors and app routes.
export function selectStartupView(
  pathname: string,
  hash: string,
  interacted: boolean,
): StartupView {
  if (pathname !== '/' && isStaticRoute(pathname) && !hash) return 'static'
  return shouldShowStartupLoading(pathname, hash, interacted)
    ? 'loading'
    : 'landing'
}

// shouldShowStartupLoading reports whether startup shows the loading screen
// instead of prerendered page content.
export function shouldShowStartupLoading(
  pathname: string,
  hash: string,
  interacted: boolean,
): boolean {
  if (hash.length > 1) return true
  if (interacted) return true
  return isPathnameAppRoute(pathname)
}
