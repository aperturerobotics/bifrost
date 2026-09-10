import { useCallback, useEffect, useRef, useState } from 'react'

import type { BillingConsent } from '@s4wave/sdk/provider/spacewave/spacewave.pb.js'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from '@s4wave/web/ui/dialog.js'
import { CLOUD_OFFER } from './pricing.js'

// BillingConsentForm binds the visible terms and spending choice to one action.
export function BillingConsentForm({
  onAccept,
  initialLimit = CLOUD_OFFER.defaultOverageLimitCents,
  spendingOnly = false,
  disabled = false,
}: {
  onAccept: (consent: BillingConsent) => void
  initialLimit?: number
  spendingOnly?: boolean
  disabled?: boolean
}) {
  const [limit, setLimit] = useState(initialLimit)
  const monthlyPrice = CLOUD_OFFER.monthlyPriceCents / 100
  const label = spendingOnly
    ? 'Save extra-usage maximum'
    : `Subscribe for $${monthlyPrice}/month`
  return (
    <div className="space-y-4 text-sm">
      <p className="text-foreground-alt text-xs">
        Extra writes: ${(CLOUD_OFFER.writeMicrodollars / 100).toFixed(2)} per
        10,000. Extra uncached reads: $
        {(CLOUD_OFFER.readMicrodollars / 100).toFixed(2)} per 10,000. Charges
        accrue proportionally.
      </p>
      <label className="grid gap-2">
        Extra-usage cap
        <select
          className="bg-background rounded border p-2"
          value={limit}
          disabled={disabled}
          onChange={(event) => setLimit(Number(event.target.value))}
        >
          {CLOUD_OFFER.overageLimitsCents.map((value) => (
            <option key={value} value={value}>
              {value === 0 ? 'Off' : `$${value / 100}/month`}
            </option>
          ))}
        </select>
      </label>
      {!spendingOnly && (
        <p>
          Your subscription renews automatically for ${monthlyPrice}/month
          before tax until you cancel online in Billing, effective at the paid
          period's end. By subscribing, you agree to the{' '}
          <a className="underline" href="#/tos">
            Terms
          </a>
          , acknowledge the{' '}
          <a className="underline" href="#/privacy">
            Privacy Policy
          </a>
          , and authorize the selected extra-usage maximum at the rates above.
        </p>
      )}
      {spendingOnly && (
        <p>
          By selecting “{label}”, you authorize the selected recurring maximum
          at the rates above. Your base subscription stays ${monthlyPrice}
          /month. Lowering the maximum preserves accrued charges and reserved
          work.
        </p>
      )}
      <button
        className="bg-brand text-background w-full rounded px-4 py-2 disabled:opacity-50"
        disabled={disabled}
        onClick={() =>
          onAccept({
            offerVersion: CLOUD_OFFER.version,
            policyVersion: CLOUD_OFFER.policyVersion,
            renewalAccepted: !spendingOnly,
            overageAccepted: limit > 0,
            overageLimitCents: limit,
          })
        }
      >
        {label}
      </button>
    </div>
  )
}

// useBillingConsent waits for the explicit action on the displayed offer.
// Dismissal and unmount cancel the pending request without granting agreement.
export function useBillingConsent() {
  const pending = useRef<((consent?: BillingConsent) => void) | null>(null)
  const [choice, setChoice] = useState<{
    limit: number
    spendingOnly: boolean
  } | null>(null)

  const finish = useCallback((consent?: BillingConsent) => {
    const resolve = pending.current
    pending.current = null
    setChoice(null)
    resolve?.(consent)
  }, [])

  useEffect(
    () => () => {
      pending.current?.()
      pending.current = null
    },
    [],
  )

  const requestConsent = useCallback(
    (
      currentLimit = CLOUD_OFFER.defaultOverageLimitCents,
      spendingOnly = false,
    ) => {
      pending.current?.()
      setChoice({ limit: currentLimit, spendingOnly })
      return new Promise<BillingConsent | undefined>((resolve) => {
        pending.current = resolve
      })
    },
    [],
  )

  const consentDialog = (
    <Dialog
      open={choice !== null}
      onOpenChange={(value) => {
        if (!value) finish()
      }}
    >
      <DialogContent className="max-h-[90dvh] overflow-y-auto">
        <DialogTitle>
          {choice?.spendingOnly ? 'Extra-usage maximum' : 'Cloud monthly offer'}
        </DialogTitle>
        <DialogDescription>
          ${CLOUD_OFFER.monthlyPriceCents / 100}/month before tax. Includes{' '}
          {CLOUD_OFFER.storageBytes / 2 ** 30} GiB of encrypted cloud storage,{' '}
          {CLOUD_OFFER.writeOperations.toLocaleString()} cloud writes, and{' '}
          {CLOUD_OFFER.readOperations.toLocaleString()} uncached cloud reads per
          subscription month.
        </DialogDescription>
        {choice && (
          <BillingConsentForm
            initialLimit={choice.limit}
            spendingOnly={choice.spendingOnly}
            onAccept={finish}
          />
        )}
      </DialogContent>
    </Dialog>
  )
  return { requestConsent, consentDialog }
}
