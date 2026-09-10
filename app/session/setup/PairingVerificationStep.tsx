import { useEffect, useEffectEvent, useState } from 'react'
import { LuCircleCheck, LuShieldCheck, LuX } from 'react-icons/lu'
import { useRetryWithAbort } from '@aptre/bldr-react'

import {
  PairingStatus,
  type WatchPairingStatusResponse,
} from '@s4wave/sdk/session/session.pb.js'
import type { Session } from '@s4wave/sdk/session/session.js'
import { pairingStatusIsTerminalFailure } from '@s4wave/app/loading/status/pairing.js'
import { Spinner } from '@s4wave/web/ui/loading/Spinner.js'
import { cn } from '@s4wave/web/style/utils.js'
import { PairingChannelProgress } from './PairingChannelProgress.js'

interface PairingVerificationStepProps {
  session: Session | null | undefined
  onContinue: () => void
  onAbort: () => void
}

// PairingVerificationStep presents the selected account and follows the provider's
// bilateral approval and enrollment operation across every pairing entry point.
export function PairingVerificationStep({
  session,
  onContinue,
  onAbort,
}: PairingVerificationStepProps) {
  const [snapshot, setSnapshot] = useState<WatchPairingStatusResponse>()
  const [submitted, setSubmitted] = useState(false)
  const [error, setError] = useState<string>()
  const continuePairing = useEffectEvent(onContinue)

  useRetryWithAbort(
    async (signal) => {
      if (!session) return
      try {
        for await (const next of session.watchPairingStatus(signal)) {
          if (signal.aborted) return
          setSnapshot(next)
          if (pairingStatusIsTerminalFailure(next.status)) {
            setError(next.errorMessage || 'Pairing failed')
            return
          }
          if (next.status === PairingStatus.PairingStatus_BOTH_CONFIRMED) {
            return
          }
        }
      } catch (cause) {
        if (!signal.aborted) {
          setError(
            cause instanceof Error
              ? cause.message
              : 'Pairing connection closed',
          )
        }
      }
    },
    undefined,
    [session],
  )

  useEffect(() => {
    if (snapshot?.status === PairingStatus.PairingStatus_BOTH_CONFIRMED) {
      continuePairing()
    }
  }, [snapshot?.status])

  const confirm = async (approved: boolean) => {
    if (!session || submitted) return
    setSubmitted(true)
    try {
      await session.confirmSASMatch(approved)
      if (!approved) onAbort()
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Confirmation failed')
      setSubmitted(false)
    }
  }

  if (error) {
    return (
      <div className="space-y-4 text-center">
        <h2 className="text-foreground text-sm font-medium">Pairing failed</h2>
        <p className="text-destructive text-xs">{error}</p>
        <button
          onClick={onAbort}
          className="border-foreground/20 text-foreground h-10 w-full rounded-md border text-sm"
        >
          Try again
        </button>
      </div>
    )
  }

  const enrolling = snapshot?.status === PairingStatus.PairingStatus_ENROLLING
  const waiting =
    submitted ||
    snapshot?.status === PairingStatus.PairingStatus_WAITING_FOR_REMOTE_CONFIRM
  const emoji = snapshot?.emoji ?? []
  const accountName = snapshot?.accountName || 'this account'

  return (
    <div className="space-y-4">
      <div className="flex flex-col items-center gap-2 text-center">
        <div className="bg-brand/10 flex size-10 items-center justify-center rounded-full">
          <LuShieldCheck className="text-brand size-5" />
        </div>
        <h2 className="text-foreground text-sm font-medium">
          {enrolling
            ? 'Connecting account'
            : waiting
              ? 'Waiting for other device'
              : 'Verify connection'}
        </h2>
        <p className="text-foreground-alt text-xs leading-relaxed">
          {enrolling
            ? 'Saving account access and preparing your Spaces.'
            : waiting
              ? 'Confirm the same emoji on the other device.'
              : snapshot?.receiving
                ? `Add ${accountName} to this device? Confirm the emoji match on both devices.`
                : `Allow the other device to access ${accountName}? Confirm the emoji match on both devices.`}
        </p>
      </div>

      {emoji.length > 0 && (
        <div
          className="grid grid-cols-3 gap-2 px-4"
          aria-label="Verification emoji"
        >
          {emoji.map((symbol, index) => (
            <div
              key={`${index}-${symbol}`}
              className="bg-foreground/5 flex h-14 items-center justify-center rounded-md text-3xl"
            >
              {symbol}
            </div>
          ))}
        </div>
      )}
      {enrolling || waiting ? (
        <div className="flex h-10 justify-center">
          <Spinner size="lg" />
        </div>
      ) : emoji.length === 0 ? (
        <PairingChannelProgress status={snapshot?.status} />
      ) : (
        <div className="flex gap-2">
          <button
            onClick={() => void confirm(false)}
            className={cn(
              'border-destructive/30 hover:bg-destructive/10 flex h-10 flex-1 items-center justify-center gap-2 rounded-md border',
            )}
          >
            <LuX className="text-destructive size-4" />
            <span className="text-destructive text-sm">No, abort</span>
          </button>
          <button
            onClick={() => void confirm(true)}
            className={cn(
              'border-brand/30 bg-brand/10 hover:bg-brand/20 flex h-10 flex-1 items-center justify-center gap-2 rounded-md border',
            )}
          >
            <LuCircleCheck className="text-brand size-4" />
            <span className="text-foreground text-sm">Yes, they match</span>
          </button>
        </div>
      )}
    </div>
  )
}
