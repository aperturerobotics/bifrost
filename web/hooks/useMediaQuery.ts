import { useCallback, useSyncExternalStore } from 'react'

// useMediaQuery reports whether the CSS media query matches, and follows its
// changes. It reports false where there is no window.
export function useMediaQuery(query: string): boolean {
  const subscribe = useCallback(
    (listener: () => void) => {
      const media = window.matchMedia(query)
      media.addEventListener('change', listener)
      return () => media.removeEventListener('change', listener)
    },
    [query],
  )
  const getSnapshot = useCallback(
    () => window.matchMedia(query).matches,
    [query],
  )
  return useSyncExternalStore(subscribe, getSnapshot, () => false)
}

// coarsePointerQuery matches a touch screen or another imprecise pointer,
// which the coarse pointer rules in web/style/app.css size targets for.
export const coarsePointerQuery = '(pointer: coarse)'
