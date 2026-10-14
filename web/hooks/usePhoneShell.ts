import { useMediaQuery } from './useMediaQuery.js'

// PHONE_SHELL_QUERY matches the phone shell rules in web/style/app.css.
const PHONE_SHELL_QUERY =
  '(max-width: 640px) and (pointer: coarse), (max-height: 470px) and (hover: none) and (pointer: coarse)'

// usePhoneShell reports whether the shell uses its phone layout.
export function usePhoneShell(): boolean {
  return useMediaQuery(PHONE_SHELL_QUERY)
}
