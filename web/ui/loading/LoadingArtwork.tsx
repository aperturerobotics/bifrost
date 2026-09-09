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

// mountLoadingArtwork paints visible hosts before the browser can present an
// empty canvas, and releases the scene when its host leaves the document.
export function mountLoadingArtwork(host: HTMLElement): () => void {
  const canvas = host.querySelector<HTMLCanvasElement>('canvas')!
  const shell = host.closest<HTMLElement>('.swl-canvas')!
  if (typeof WebGL2RenderingContext === 'undefined') return () => {}
  let release: (() => void) | undefined
  function show(visible: boolean) {
    if (!visible) {
      release?.()
      release = undefined
      return
    }
    if (release) return
    try {
      release = createLoadingArtwork(canvas, shell)
    } catch {
      // Status and recovery remain usable when a GPU context is unavailable.
      canvas.dataset.renderer = 'unavailable'
    }
  }
  const visibility = new IntersectionObserver(([entry]) => {
    show(entry.isIntersecting)
  })
  const bounds = canvas.getBoundingClientRect()
  show(
    bounds.width > 0 &&
      bounds.height > 0 &&
      bounds.bottom > 0 &&
      bounds.right > 0 &&
      bounds.top < window.innerHeight &&
      bounds.left < window.innerWidth,
  )
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
