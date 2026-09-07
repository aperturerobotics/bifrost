import { useCallback } from 'react'

import { createLoadingArtwork } from './loading-artwork.js'

// LoadingArtwork attaches the decorative scene to its DOM lifetime, including
// static startup HTML that is later replaced by the application root.
export function LoadingArtwork() {
  const ref = useCallback((element: HTMLDivElement | null) => {
    if (!element) return
    return mountLoadingArtwork(element)
  }, [])
  return (
    <div ref={ref} className="swl-artwork-host">
      <canvas className="swl-artwork" />
    </div>
  )
}

// mountLoadingArtwork releases the scene when its host leaves the document.
export function mountLoadingArtwork(host: HTMLElement): () => void {
  const canvas = host.querySelector<HTMLCanvasElement>('canvas')!
  const shell = host.closest<HTMLElement>('.swl-canvas')!
  if (typeof WebGL2RenderingContext === 'undefined') return () => {}
  let release: (() => void) | undefined
  const visibility = new IntersectionObserver(([entry]) => {
    release?.()
    release = undefined
    if (!entry.isIntersecting) return
    try {
      release = createLoadingArtwork(canvas, shell)
    } catch {
      // Status and recovery remain usable when a GPU context is unavailable.
      canvas.dataset.renderer = 'unavailable'
    }
  })
  visibility.observe(canvas)
  const removal = new MutationObserver(() => {
    if (!host.isConnected) dispose()
  })
  removal.observe(document.body, { childList: true, subtree: true })
  function dispose() {
    visibility.disconnect()
    removal.disconnect()
    release?.()
    release = undefined
  }
  return dispose
}
