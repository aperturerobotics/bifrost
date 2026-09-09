import { useMemo } from 'react'

import { useSessionIndex } from '@s4wave/web/contexts/SessionIndexContext.js'
import {
  Route,
  RouterContext,
  Routes,
  useRouter,
} from '@s4wave/web/router/router.js'
import { NavigatePath } from '@s4wave/web/router/NavigatePath.js'

import { BlogRoutes } from './BlogRoutes.js'
import { DocsRoutes } from './DocsRoutes.js'
import { LandingRoutes } from './LandingRoutes.js'
import { usePageNavigate } from './usePageNavigate.js'

// SessionPages reuses public pages beneath the mounted session and its footer.
export function SessionPages() {
  const router = useRouter()
  const sessionIndex = useSessionIndex()
  const navigate = usePageNavigate()
  const context = useMemo(() => ({ ...router, navigate }), [router, navigate])
  const path = router.path
    .slice(`/u/${sessionIndex}`.length)
    .replace(/^\/legal\/?$/, '/tos')
    .replace(/^\/legal\//, '/')

  return (
    <RouterContext value={context}>
      <Routes path={path} fullPath>
        {LandingRoutes}
        {DocsRoutes}
        {BlogRoutes}
        <Route path="*">
          <NavigatePath to={`/u/${sessionIndex}`} replace />
        </Route>
      </Routes>
    </RouterContext>
  )
}
