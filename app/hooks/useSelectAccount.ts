import { useCallback } from 'react'

import { useAppNavigation } from '@s4wave/web/sdk/app/environment.js'

// useSelectAccount returns the shared account-selection navigation action.
export function useSelectAccount(): (sessionIndex: number) => void {
  const { setAppPath } = useAppNavigation()
  return useCallback(
    (sessionIndex: number) => {
      setAppPath(`/u/${sessionIndex}/`)
    },
    [setAppPath],
  )
}
