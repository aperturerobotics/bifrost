import { lazy } from 'react'

import { Route, useParams } from '@s4wave/web/router/router.js'
import { CheckoutResultPage } from '@s4wave/app/provider/spacewave/CheckoutResultPage.js'
import { useSessionList } from '@s4wave/app/hooks/useSessionList.js'
import { NavigatePath } from '@s4wave/web/router/NavigatePath.js'
import { LoadingCard } from '@s4wave/web/ui/loading/LoadingCard.js'

import { formatPendingJoin, storePendingJoin } from './pendingJoin.js'

export { consumePendingJoin, storePendingJoin } from './pendingJoin.js'

const LazyPairCodePage = lazy(async () => {
  const { PairCodePage } = await import('@s4wave/app/pair/PairCodePage.js')
  return { default: PairCodePage }
})

// JoinRedirect resolves the first available session and redirects to its join route.
function JoinRedirect() {
  const params = useParams()
  const code = params.code ?? ''
  const resource = useSessionList()

  if (resource.loading) {
    return (
      <div className="flex h-full w-full items-center justify-center p-6">
        <div className="w-full max-w-sm">
          <LoadingCard
            view={{
              state: 'loading',
              title: 'Preparing join',
              detail: 'Resolving your session before redirecting.',
            }}
          />
        </div>
      </div>
    )
  }

  const sessions = resource.value?.sessions ?? []
  if (sessions.length === 0) {
    // Stash the invite code so it survives account creation.
    if (code) storePendingJoin(code)
    return <NavigatePath to="/" replace />
  }

  const idx = sessions[0].sessionIndex ?? 1
  const target = code
    ? `/u/${idx}/join/${formatPendingJoin(code)}`
    : `/u/${idx}/join`
  return <NavigatePath to={target} replace />
}

// SessionRoutes contains checkout, join, and pairing entry routes.
// AppRoutes owns Quickstart and the mounted session surface.
export const SessionRoutes = (
  <>
    <Route path="/checkout/success">
      <CheckoutResultPage success />
    </Route>
    <Route path="/checkout/cancel">
      <CheckoutResultPage />
    </Route>
    <Route path="/join/:code">
      <JoinRedirect />
    </Route>
    <Route path="/join">
      <JoinRedirect />
    </Route>
    <Route path="/pair/:code">
      <LazyPairCodePage />
    </Route>
    <Route path="/pair">
      <LazyPairCodePage />
    </Route>
  </>
)
