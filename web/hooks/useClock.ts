import { useEffect, useState } from 'react'

// useClock returns the wall clock in milliseconds, advanced every periodMs.
// Use it for text and states that change with time alone, such as "scraped
// 2 minutes ago" or a deadline passing while the watched source is silent.
export function useClock(periodMs: number): number {
  const [now, setNow] = useState(Date.now)
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), periodMs)
    return () => clearInterval(id)
  }, [periodMs])
  return now
}
