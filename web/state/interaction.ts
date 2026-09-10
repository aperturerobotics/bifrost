const INTERACTION_KEY = 'spacewave-has-interacted'

export function hasInteracted(storage = localStorage): boolean {
  return storage.getItem(INTERACTION_KEY) === 'true'
}

export function markInteracted(storage = localStorage): void {
  storage.setItem(INTERACTION_KEY, 'true')
}

export function clearInteracted(storage = localStorage): void {
  storage.removeItem(INTERACTION_KEY)
}
