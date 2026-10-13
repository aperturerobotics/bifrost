import { createContext, use, useEffect } from 'react'
import type { ObjectViewerComponent } from './object.js'

export interface ObjectViewerContextValue {
  visibleComponents: ObjectViewerComponent[]
  selectedComponent?: ObjectViewerComponent
  onSelectComponent: (component: ObjectViewerComponent) => void
  // setObjectName reports the stored name of the shown object.
  setObjectName: (name: string) => void
}

const ObjectViewerContext = createContext<ObjectViewerContextValue | null>(null)

export const ObjectViewerProvider = ObjectViewerContext.Provider

// useObjectViewer returns the viewer context or null when outside a provider.
export function useObjectViewer(): ObjectViewerContextValue | null {
  return use(ObjectViewerContext)
}

// useObjectName reports the stored name of the object a viewer shows, so its
// tab and the document title use the name in place of the object key. An
// empty name reports nothing.
export function useObjectName(name: string | undefined) {
  const setObjectName = use(ObjectViewerContext)?.setObjectName
  useEffect(() => {
    if (name) setObjectName?.(name)
  }, [setObjectName, name])
}
