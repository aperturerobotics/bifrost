import { useCallback, useState } from 'react'

import { useStateAtom, useStateNamespace } from '@s4wave/web/state/index.js'

import type { NotebookSource } from './proto/notebook.pb.js'
import type { NotebookHandle } from './sdk/notebook.js'

interface UseNotebookViewerStateOptions {
  sources: NotebookSource[]
  notebookHandle: NotebookHandle | null | undefined
}

// useNotebookViewerState owns notebook selection, filtering, responsive
// navigation, and source mutation transitions for the viewer.
export function useNotebookViewerState({
  sources,
  notebookHandle,
}: UseNotebookViewerStateOptions) {
  const ns = useStateNamespace(['notes'])

  // Persisted state for selected source and note.
  const [selectedSource, setSelectedSource] = useStateAtom<number>(
    ns,
    'selectedSource',
    0,
  )
  const [selectedNote, setSelectedNote] = useStateAtom<string>(
    ns,
    'selectedNote',
    '',
  )
  const [currentPath, setCurrentPath] = useStateAtom<string>(
    ns,
    'currentPath',
    '',
  )
  const [editing, setEditing] = useStateAtom<boolean>(ns, 'editing', false)

  // Tag filter state.
  const [filterTag, setFilterTag] = useState<string | undefined>(undefined)
  const [filterStatus, setFilterStatus] = useState<string | undefined>(
    undefined,
  )

  // Responsive sidebar visibility.
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [addSourceOpen, setAddSourceOpen] = useState(false)
  const [removeSourceIndex, setRemoveSourceIndex] = useState<number | null>(
    null,
  )

  const currentSource = sources[selectedSource]

  const handleSelectSource = useCallback(
    (index: number) => {
      setSelectedSource(index)
      setCurrentPath('')
      setSelectedNote('')
      setEditing(false)
      setFilterTag(undefined)
      setFilterStatus(undefined)
    },
    [setSelectedSource, setCurrentPath, setSelectedNote, setEditing],
  )

  const handleAddSource = useCallback(() => {
    setAddSourceOpen(true)
  }, [])

  const handleConfirmAddSource = useCallback(
    async ({ name, ref }: { name: string; ref: string }) => {
      if (!notebookHandle) return
      if (!ref) return

      await notebookHandle.addSource({ name, ref })
      setAddSourceOpen(false)
      setSelectedSource(sources.length)
      setCurrentPath('')
      setSelectedNote('')
      setEditing(false)
      setFilterTag(undefined)
      setFilterStatus(undefined)
    },
    [
      notebookHandle,
      setSelectedSource,
      setCurrentPath,
      setSelectedNote,
      setEditing,
      sources.length,
    ],
  )

  const handleRemoveSource = useCallback((index: number) => {
    setRemoveSourceIndex(index)
  }, [])

  const handleConfirmRemoveSource = useCallback(async () => {
    if (!notebookHandle || removeSourceIndex === null) return

    const index = removeSourceIndex
    await notebookHandle.removeSource(index)
    setRemoveSourceIndex(null)
    setSelectedSource((prev) => {
      if (prev > index) return prev - 1
      if (prev === index) return Math.max(0, prev - 1)
      return prev
    })
    setCurrentPath('')
    setSelectedNote('')
    setEditing(false)
    setFilterTag(undefined)
    setFilterStatus(undefined)
  }, [
    notebookHandle,
    removeSourceIndex,
    setSelectedSource,
    setCurrentPath,
    setSelectedNote,
    setEditing,
  ])

  const handleMoveSource = useCallback(
    async (index: number, delta: -1 | 1) => {
      if (!notebookHandle) return
      const nextIndex = index + delta
      if (nextIndex < 0 || nextIndex >= sources.length) return

      const order = sources.map((_, idx) => idx)
      ;[order[index], order[nextIndex]] = [order[nextIndex], order[index]]
      await notebookHandle.reorderSources(order)
      setSelectedSource((prev) => {
        if (prev === index) return nextIndex
        if (prev === nextIndex) return index
        return prev
      })
    },
    [notebookHandle, sources, setSelectedSource],
  )

  const handleSelectNote = useCallback(
    (path: string) => {
      setCurrentPath(getParentPath(path))
      setSelectedNote(path)
      setEditing(false)
      setSidebarOpen(false)
    },
    [setCurrentPath, setSelectedNote, setEditing],
  )

  const handleChangePath = useCallback(
    (path: string) => {
      setCurrentPath(path)
      setSelectedNote('')
      setEditing(false)
    },
    [setCurrentPath, setSelectedNote, setEditing],
  )

  const handleNoteRenamed = useCallback(
    (prevPath: string, nextPath: string) => {
      if (selectedNote === prevPath) {
        setSelectedNote(nextPath)
      }
    },
    [selectedNote, setSelectedNote],
  )

  const handleNoteDeleted = useCallback(
    (path: string) => {
      if (selectedNote !== path) return
      setSelectedNote('')
      setEditing(false)
    },
    [selectedNote, setSelectedNote, setEditing],
  )

  const handleToggleEdit = useCallback(() => {
    setEditing((prev) => !prev)
  }, [setEditing])

  return {
    ns,
    selectedSource,
    selectedNote,
    currentPath,
    editing,
    filterTag,
    filterStatus,
    sidebarOpen,
    addSourceOpen,
    removeSourceIndex,
    currentSource,
    setFilterTag,
    setFilterStatus,
    setSidebarOpen,
    setAddSourceOpen,
    setRemoveSourceIndex,
    handleSelectSource,
    handleAddSource,
    handleConfirmAddSource,
    handleRemoveSource,
    handleConfirmRemoveSource,
    handleMoveSource,
    handleSelectNote,
    handleChangePath,
    handleNoteRenamed,
    handleNoteDeleted,
    handleToggleEdit,
  }
}

function getParentPath(path: string): string {
  const parts = path.split('/').filter(Boolean)
  parts.pop()
  return parts.join('/')
}
