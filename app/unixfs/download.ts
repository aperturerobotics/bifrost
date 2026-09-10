import { pluginPathPrefix } from '@s4wave/app/urls.js'
import { ExportBatchRequest } from '@s4wave/core/space/http/export/config.pb.js'
import {
  buildProjectedExportURL,
  buildProjectedFileInlineURL,
  buildProjectedFileURL,
  buildProjectedObjectContentPath,
} from '@s4wave/app/space/projected-url.js'
import { joinProjectedSubpath } from '@s4wave/sdk/space/projected-path.js'
import { downloadURL } from '@s4wave/web/download.js'
import type { DownloadDragTarget } from '@s4wave/web/dnd/download-url-drag.js'
import type { FileEntry } from '@s4wave/web/editors/file-browser/types.js'

interface UnixFSSelectionDownloadOpts {
  httpPathPrefix?: string
  sessionIndex: number
  sharedObjectId: string
  objectKey: string
  currentPath: string
  entries: FileEntry[]
}

type UnixFSSelectionDownloadDragTargetOpts = UnixFSSelectionDownloadOpts

function buildBatchFilename(entries: FileEntry[]): string {
  if (entries.length === 1) {
    return `${entries[0].name}.zip`
  }
  return 'selection.zip'
}

function normalizeBatchEntries(entries: FileEntry[]): FileEntry[] {
  const seen = new Set<string>()
  return [...entries]
    .filter((entry) => {
      if (seen.has(entry.name)) {
        return false
      }
      seen.add(entry.name)
      return true
    })
    .sort((a, b) => a.name.localeCompare(b.name))
}

function encodeBase64Url(data: Uint8Array): string {
  let binary = ''
  for (let i = 0; i < data.length; i += 0x8000) {
    binary += String.fromCharCode(...data.subarray(i, i + 0x8000))
  }
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

export function buildUnixFSFileDownloadURL(
  sessionIndex: number,
  sharedObjectId: string,
  objectKey: string,
  path: string,
  httpPathPrefix = '',
): string {
  return buildProjectedFileURL({
    sessionIndex,
    sharedObjectId,
    objectKey,
    path,
    httpPathPrefix,
  })
}

export function buildUnixFSFileInlineURL(
  sessionIndex: number,
  sharedObjectId: string,
  objectKey: string,
  path: string,
  httpPathPrefix = '',
): string {
  return buildProjectedFileInlineURL({
    sessionIndex,
    sharedObjectId,
    objectKey,
    path,
    httpPathPrefix,
  })
}

export function buildUnixFSExportURL(
  sessionIndex: number,
  sharedObjectId: string,
  objectKey: string,
  path: string,
  httpPathPrefix = '',
): string {
  return buildProjectedExportURL({
    sessionIndex,
    sharedObjectId,
    objectKey,
    path,
    httpPathPrefix,
  })
}

export function buildUnixFSBatchExportURL(
  sessionIndex: number,
  sharedObjectId: string,
  objectKey: string,
  basePath: string,
  entries: FileEntry[],
  httpPathPrefix = '',
): { url: string; filename: string } {
  const normalizedEntries = normalizeBatchEntries(entries)
  const req = ExportBatchRequest.toBinary({
    paths: normalizedEntries.map((entry) => entry.name),
  })
  const baseProjectedPath = buildProjectedObjectContentPath({
    sessionIndex,
    sharedObjectId,
    objectKey,
    path: basePath,
  })
  const filename = buildBatchFilename(normalizedEntries)
  const encodedFilename = encodeURIComponent(filename)
  const encodedReq = encodeBase64Url(req)
  return {
    url:
      `${pluginPathPrefix}${httpPathPrefix}/export-batch/${baseProjectedPath}/` +
      `${encodedReq}/${encodedFilename}`,
    filename,
  }
}

export function buildUnixFSSelectionDownloadDragTarget({
  httpPathPrefix,
  sessionIndex,
  sharedObjectId,
  objectKey,
  currentPath,
  entries,
}: UnixFSSelectionDownloadDragTargetOpts): DownloadDragTarget | null {
  const normalizedEntries = normalizeBatchEntries(entries)
  if (normalizedEntries.length === 0) {
    return null
  }

  if (normalizedEntries.length === 1 && !normalizedEntries[0].isDir) {
    const filePath = joinProjectedSubpath([
      currentPath,
      normalizedEntries[0].name,
    ])
    return {
      mimeType: 'application/octet-stream',
      filename: normalizedEntries[0].name,
      url: buildUnixFSFileDownloadURL(
        sessionIndex,
        sharedObjectId,
        objectKey,
        filePath,
        httpPathPrefix,
      ),
    }
  }

  if (normalizedEntries.length === 1) {
    const dirPath = joinProjectedSubpath([
      currentPath,
      normalizedEntries[0].name,
    ])
    return {
      mimeType: 'application/zip',
      filename: `${normalizedEntries[0].name}.zip`,
      url: buildUnixFSExportURL(
        sessionIndex,
        sharedObjectId,
        objectKey,
        dirPath,
        httpPathPrefix,
      ),
    }
  }

  const batchDownload = buildUnixFSBatchExportURL(
    sessionIndex,
    sharedObjectId,
    objectKey,
    currentPath,
    normalizedEntries,
    httpPathPrefix,
  )
  return {
    mimeType: 'application/zip',
    filename: batchDownload.filename,
    url: batchDownload.url,
  }
}

export function downloadUnixFSSelection({
  httpPathPrefix,
  sessionIndex,
  sharedObjectId,
  objectKey,
  currentPath,
  entries,
}: UnixFSSelectionDownloadOpts): Promise<void> {
  const normalizedEntries = normalizeBatchEntries(entries)
  if (normalizedEntries.length === 0) {
    return Promise.resolve()
  }

  if (normalizedEntries.length === 1 && !normalizedEntries[0].isDir) {
    const filePath = joinProjectedSubpath([
      currentPath,
      normalizedEntries[0].name,
    ])
    return downloadURL(
      buildUnixFSFileDownloadURL(
        sessionIndex,
        sharedObjectId,
        objectKey,
        filePath,
        httpPathPrefix,
      ),
      normalizedEntries[0].name,
    )
  }

  if (normalizedEntries.length === 1) {
    const dirPath = joinProjectedSubpath([
      currentPath,
      normalizedEntries[0].name,
    ])
    return downloadURL(
      buildUnixFSExportURL(
        sessionIndex,
        sharedObjectId,
        objectKey,
        dirPath,
        httpPathPrefix,
      ),
      `${normalizedEntries[0].name}.zip`,
    )
  }

  const batchDownload = buildUnixFSBatchExportURL(
    sessionIndex,
    sharedObjectId,
    objectKey,
    currentPath,
    normalizedEntries,
    httpPathPrefix,
  )
  return downloadURL(batchDownload.url, batchDownload.filename)
}
