import { useCallback } from 'react'

import { useNavigate, useParams } from '@s4wave/web/router/router.js'
import { usePromise } from '@s4wave/web/hooks/usePromise.js'
import { Quickstart } from '@s4wave/app/quickstart/Quickstart.js'
import { QuickstartUnavailable } from '@s4wave/app/quickstart/QuickstartUnavailable.js'
import { isQuickstartId } from '@s4wave/app/quickstart/options.js'

import './AppQuickstart.css'

// AppQuickstart creates a workspace while loading its destination module.
export function AppQuickstart() {
  const quickstartId = useParams()['quickstartId']
  const navigate = useNavigate()
  usePromise(
    useCallback(() => {
      if (quickstartId && isQuickstartId(quickstartId)) {
        return import('./AppSession.js')
      }
      return undefined
    }, [quickstartId]),
  )
  const navigateHome = useCallback(() => {
    navigate({ path: '/' })
  }, [navigate])
  if (!quickstartId || !isQuickstartId(quickstartId)) {
    return (
      <QuickstartUnavailable
        quickstartId={quickstartId}
        onBack={navigateHome}
      />
    )
  }

  return <Quickstart quickstartId={quickstartId} />
}
