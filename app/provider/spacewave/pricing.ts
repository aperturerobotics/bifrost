import cloudOffer from '@s4wave/core/provider/spacewave/api/cloud-offer.json'

export const CLOUD_OFFER = cloudOffer
export const PLAN_PRICE_MONTHLY = CLOUD_OFFER.monthlyPriceCents / 100
export const STORAGE_BASELINE_GB = CLOUD_OFFER.storageBytes / 2 ** 30
export const WRITE_OPS_BASELINE = CLOUD_OFFER.writeOperations
export const WRITE_OPS_BASELINE_DISPLAY = `${WRITE_OPS_BASELINE / 1000}K`
export const READ_OPS_BASELINE = CLOUD_OFFER.readOperations
export const READ_OPS_BASELINE_DISPLAY = `${READ_OPS_BASELINE / 1000}K`
export const OVERAGE_WRITE_PER_MILLION = CLOUD_OFFER.writeMicrodollars
export const OVERAGE_READ_PER_MILLION = CLOUD_OFFER.readMicrodollars
export const OVERAGE_EXPLANATION = `Extra usage starts with a $10 monthly maximum, which you can change or turn off. Extra writes cost $${(CLOUD_OFFER.writeMicrodollars / 100).toFixed(2)} per 10,000 and extra uncached reads cost $${(CLOUD_OFFER.readMicrodollars / 100).toFixed(2)} per 10,000, accrued proportionally. Choose a $5, $10, or $20 monthly maximum. Allowances reset at your subscription renewal. Storage is limited to ${STORAGE_BASELINE_GB} GiB.`

export const FREE_FEATURES = [
  'Full local-first app on your devices',
  'No account, sign-up, or payment required',
  'Stores data on your devices',
  'Peer-to-peer sync directly between devices',
  'Full plugin SDK and developer tools',
  'End-to-end encrypted by default',
  'Open-source, self-hostable',
]

export const CLOUD_FEATURES = [
  'Adds cloud sync, storage, backup, and relay services:',
  'Cloud sync and backup across all devices',
  'Shared Spaces with collaborators',
  `${STORAGE_BASELINE_GB} GiB cloud storage included`,
  `${WRITE_OPS_BASELINE_DISPLAY} write operations / month`,
  `${READ_OPS_BASELINE_DISPLAY} cloud reads / month`,
  'Extra usage with a $10 monthly maximum you control',
]

export interface OverageItem {
  resource: string
  baseline: string
  rate: string
}

export const OVERAGE_ITEMS: OverageItem[] = [
  {
    resource: 'Write operations',
    baseline: `${WRITE_OPS_BASELINE_DISPLAY} / month`,
    rate: `$${(OVERAGE_WRITE_PER_MILLION / 100).toFixed(2)} / 10,000`,
  },
  {
    resource: 'Cloud reads',
    baseline: `${READ_OPS_BASELINE_DISPLAY} / month`,
    rate: `$${(OVERAGE_READ_PER_MILLION / 100).toFixed(2)} / 10,000`,
  },
]
