import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { UsageBars } from './UsageBars.js'

const mockBillingState = vi.hoisted(() => ({
  selfServiceAllowed: false,
  response: {
    usage: {
      storageBytes: 80 * 2 ** 30,
      storageBaselineBytes: 100 * 2 ** 30,
      writeOps: 42_500n,
      writeOpsBaseline: 50_000n,
      readOps: 225_000n,
      readOpsBaseline: 250_000n,
      overageLimitCents: 1000,
      accruedOverageMicrodollars: 2_000_000n,
      reservedOverageMicrodollars: 500_000n,
      currentPeriodStart: 1_800_000_000_000n,
      currentPeriodEnd: 1_802_592_000_000n,
    },
  },
}))

vi.mock('./BillingStateProvider.js', () => ({
  useBillingStateContext: () => mockBillingState,
}))
vi.mock('@s4wave/web/contexts/contexts.js', () => ({
  SessionContext: { useContext: () => ({ value: null }) },
}))

afterEach(cleanup)

describe('UsageBars', () => {
  it('shows accrued charges, pending exposure, available budget, and the subscription reset date', () => {
    render(<UsageBars />)
    expect(
      screen.getByText(
        /Accrued: \$2.00 · Reserved: \$0.50 · Available: \$7.50/,
      ),
    ).toBeDefined()
    expect(screen.getByText(/Subscription period:/)).toBeDefined()
    expect(screen.getByText(/Peer-only traffic and cached reads/)).toBeDefined()
    expect(screen.queryByText('Extra storage')).toBeNull()
  })

  it('shows threshold alerts against the monthly offer', () => {
    render(<UsageBars />)
    expect(
      screen.getByText('Storage has reached 80% of included usage.'),
    ).toBeDefined()
    expect(
      screen.getByText('Write Ops has reached 85% of included usage.'),
    ).toBeDefined()
    expect(
      screen.getByText('Cloud Reads has reached 90% of included usage.'),
    ).toBeDefined()
  })
})
