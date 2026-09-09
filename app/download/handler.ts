import { useCallback } from 'react'

import { usePageNavigate } from '@s4wave/app/routes/usePageNavigate.js'

// useDownloadDesktopApp opens downloads within the current session when present.
export function useDownloadDesktopApp(): () => void {
  const navigate = usePageNavigate()
  return useCallback(() => navigate({ path: '/download' }), [navigate])
}
