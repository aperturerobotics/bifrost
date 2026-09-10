import { useAppEnvironment } from '@s4wave/web/sdk/app/environment.js'
// Quickstart prepares a local session and opens the selected app.

import { useCallback, useState } from 'react'

import { Redirect } from '@s4wave/web/router/Redirect.js'
import { useRootResource } from '@s4wave/web/hooks/useRootResource.js'
import { useNavigate } from '@s4wave/web/router/router.js'
import { useResource } from '@aptre/bldr-sdk/hooks/useResource.js'
import { BackButton } from '@s4wave/web/ui/BackButton.js'
import { ErrorState } from '@s4wave/web/ui/ErrorState.js'

import {
  buildQuickstartSpaceRoutePath,
  createQuickstartSetup,
  createLocalSession,
  getQuickstartInitialObjectRouteHandoff,
  type QuickstartProgressState,
} from './create.js'
import { LoadingScreen } from './LoadingScreen.js'
import { isQuickstartCreateId, type QuickstartId } from './options.js'
import { NavigatePath } from '@s4wave/web/router/NavigatePath.js'
import { createQuickstartSessionHandoffCleanup } from './session-handoff.js'
import { markQuickstartStartupBoundary } from './startup-boundary.js'

interface QuickstartProps {
  quickstartId: QuickstartId
}

interface QuickstartErrorStateProps {
  message: string
  onRetry?: () => void
}

function QuickstartErrorState({ message, onRetry }: QuickstartErrorStateProps) {
  const navigate = useNavigate()

  return (
    <ErrorState
      variant="fullscreen"
      className="relative"
      title="Setup Failed"
      message={message}
      onRetry={onRetry}
    >
      <BackButton floating onClick={() => navigate({ path: '../../' })}>
        Back to home
      </BackButton>
    </ErrorState>
  )
}

// Quickstart retains setup resources across the final route transition.
export const Quickstart: React.FC<QuickstartProps> = ({ quickstartId }) => {
  const environment = useAppEnvironment()
  // isCreate indicates this is an option that should call createQuickstartSetup.
  // otherwise we redirect below.
  const isCreate = isQuickstartCreateId(quickstartId)
  const isLocal = quickstartId === 'local'
  const rootResource = useRootResource()
  const [progress, setProgress] = useState<QuickstartProgressState | null>(null)
  const reportProgress = useCallback((state: QuickstartProgressState) => {
    setProgress(state)
  }, [])

  // For 'local' (login page "Continue without account"), always create new session.
  const localSessionResource = useResource(
    rootResource,
    async (root, signal, cleanup) => {
      const handoff = createQuickstartSessionHandoffCleanup(
        cleanup,
        environment.instanceKey,
      )
      try {
        const setup = await createLocalSession(
          root,
          signal,
          handoff.cleanup,
          true,
          undefined,
          reportProgress,
        )
        if (signal.aborted) {
          handoff.releaseHeldResources()
          return setup
        }
        handoff.stage(setup.sessionIndex, setup.session)
        return setup
      } catch (err) {
        handoff.releaseHeldResources()
        throw err
      }
    },
    [reportProgress],
    { enabled: isLocal },
  )

  // For other create options, we create account/session/space
  const setupResource = useResource(
    rootResource,
    async (root, signal, cleanup) => {
      if (!isCreate || isLocal) return null
      const handoff = createQuickstartSessionHandoffCleanup(
        cleanup,
        environment.instanceKey,
      )
      try {
        const setup = await createQuickstartSetup(
          root,
          quickstartId,
          signal,
          handoff.cleanup,
          reportProgress,
        )
        if (signal.aborted) {
          handoff.releaseHeldResources()
          return setup
        }
        if (typeof setup.sessionIndex === 'number') {
          const spaceID =
            setup.spaceResp.sharedObjectRef?.providerResourceRef?.id
          handoff.stage(
            setup.sessionIndex,
            setup.session,
            spaceID,
            setup.initialObjectRoute ??
              getQuickstartInitialObjectRouteHandoff(quickstartId),
          )
          if (spaceID) {
            markQuickstartStartupBoundary('quickstart.route-handoff-ready', {
              quickstartId,
              sessionIndex: setup.sessionIndex,
              sharedObjectId: spaceID,
            })
          }
        }
        return setup
      } catch (err) {
        handoff.releaseHeldResources()
        throw err
      }
    },
    [isCreate, isLocal, quickstartId, reportProgress],
    { enabled: isCreate && !isLocal },
  )
  const setup = setupResource.value
  const spaceID = setup?.spaceResp.sharedObjectRef?.providerResourceRef?.id

  // account option: redirect to /login
  if (quickstartId === 'account') {
    return <NavigatePath to="/login" />
  }

  if (!isCreate) {
    // We shouldn't have gotten here
    console.error(`unknown quickstart option: ${String(quickstartId)}`)
    return <NavigatePath to="/" />
  }

  // Handle 'local' quickstart
  if (isLocal) {
    if (localSessionResource.error) {
      return (
        <QuickstartErrorState
          message={localSessionResource.error.message}
          onRetry={localSessionResource.retry}
        />
      )
    }

    const localSetup = localSessionResource.value

    if (localSessionResource.loading || !localSetup) {
      return <LoadingScreen quickstartId={quickstartId} progress={progress} />
    }

    return <Redirect to={`/u/${localSetup.sessionIndex}`} />
  }

  // Handle other create options (with space)
  if (setupResource.error) {
    return (
      <QuickstartErrorState
        message={setupResource.error.message}
        onRetry={setupResource.retry}
      />
    )
  }

  if (setupResource.loading || !setup || !spaceID) {
    return <LoadingScreen quickstartId={quickstartId} progress={progress} />
  }

  return (
    <Redirect
      to={buildQuickstartSpaceRoutePath(
        `/u/${setup.sessionIndex}/so/${spaceID}`,
        quickstartId,
        setup.initialObjectRoute?.objectKey,
      )}
    />
  )
}
