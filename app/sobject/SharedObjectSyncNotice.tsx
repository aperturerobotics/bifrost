import { useEffect, useId } from 'react'

import type { SharedObjectHealth } from '@s4wave/core/sobject/sobject.pb.js'
import { toast } from '@s4wave/web/ui/toaster.js'

// SharedObjectSyncNotice presents peer recovery without interrupting local content.
export function SharedObjectSyncNotice({
  health,
}: {
  health?: SharedObjectHealth
}) {
  const id = useId()
  const recovery = (health?.syncRecoveryPeerIds?.length ?? 0) > 0
  const denied = (health?.syncDeniedPeerIds?.length ?? 0) > 0

  useEffect(() => {
    if (!recovery && !denied) return
    toast.warning('Direct sync needs attention', {
      id,
      description: recovery
        ? 'Update both devices. If sync still cannot reconnect, ask the owner for a new invite. You can keep using local content.'
        : 'A connected device declined to sync this Space. Ask the owner to confirm your access. You can keep using local content.',
      duration: Infinity,
      closeButton: true,
    })
    return () => {
      toast.dismiss(id)
    }
  }, [denied, id, recovery])

  return null
}
