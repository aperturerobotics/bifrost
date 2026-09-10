import type { RegisterCleanup } from '@aptre/bldr-sdk/hooks/useResource.js'
import type { Root } from '@s4wave/sdk/root'
import type { Session } from '@s4wave/sdk/session/session.js'
import { LocalProvider } from '@s4wave/sdk/provider/local/local.js'

// preparePairingSession retains a private exchange identity until this view closes.
// The pairing operation registers only the Session attached to the offered account.
export async function preparePairingSession(
  root: Root,
  signal: AbortSignal,
  cleanup: RegisterCleanup,
): Promise<Session> {
  using provider = await root.lookupProvider('local', signal)
  const local = new LocalProvider(provider.resourceRef)
  const sessionRef = await local.preparePairingSession(signal)
  return cleanup(await root.mountSession({ sessionRef }, signal))
}
