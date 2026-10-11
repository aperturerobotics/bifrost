import { useStaticHref } from '@s4wave/app/prerender/StaticContext.js'

export interface TermsAgreementProps {
  agreed: boolean
  onChange: (agreed: boolean) => void
  disabled: boolean
}

// TermsAgreement renders the checkbox a new Spacewave account must tick to
// accept the Terms of Service and Privacy Policy. The links open in a new tab
// so the pending sign-up survives reading them.
export function TermsAgreement({
  agreed,
  onChange,
  disabled,
}: TermsAgreementProps) {
  const tosHref = useStaticHref('/tos')
  const privacyHref = useStaticHref('/privacy')

  return (
    <label className="flex cursor-pointer items-start gap-2 select-none">
      <input
        type="checkbox"
        checked={agreed}
        onChange={(e) => onChange(e.target.checked)}
        disabled={disabled}
        className="accent-brand mt-0.5 size-4 shrink-0 rounded"
      />
      <span className="text-foreground-alt text-xs leading-relaxed">
        I agree to the{' '}
        <a
          href={tosHref}
          target="_blank"
          rel="noopener noreferrer"
          className="text-brand hover:underline"
        >
          Terms of Service
        </a>{' '}
        and{' '}
        <a
          href={privacyHref}
          target="_blank"
          rel="noopener noreferrer"
          className="text-brand hover:underline"
        >
          Privacy Policy
        </a>
      </span>
    </label>
  )
}
