import { useCallback, useState } from 'react'

import type { ObjectViewerComponentProps } from '@s4wave/web/object/object.js'
import { getObjectKey } from '@s4wave/web/object/object.js'
import {
  useResource,
  useResourceValue,
} from '@aptre/bldr-sdk/hooks/useResource.js'
import { useStreamingResource } from '@aptre/bldr-sdk/hooks/useStreamingResource.js'
import { formatV86ImageLabel, readV86Image } from '@s4wave/app/vm/v86-images.js'
import { SpaceContentsContext } from '@s4wave/web/contexts/contexts.js'
import { LoadingCard } from '@s4wave/web/ui/loading/LoadingCard.js'
import type { SpaceContents } from '@s4wave/sdk/space/contents.js'
import { SetV86StateOp, VmState, VmV86 } from '@s4wave/sdk/vm/v86.pb.js'
import { keyToIRI, iriToKey } from '@s4wave/sdk/world/graph-utils.js'
import type { IObjectState } from '@s4wave/sdk/world/object-state.js'
import { listObjectsWithType } from '@s4wave/sdk/world/types/types.js'
import type { IWorldState } from '@s4wave/sdk/world/world-state.js'
import { UnixFSTypeID } from '@s4wave/sdk/unixfs/type.js'
import {
  PRED_V86_IMAGE,
  SET_V86_STATE_OP_ID,
  VmV86TypeID,
} from '@s4wave/sdk/vm/vmv86.js'

import { VmErrorBanner } from './VmErrorBanner.js'
import { VmHeader } from './VmHeader.js'
import { VmMachineCard, type OverrideSlot } from './VmMachineCard.js'
import { VmRunningView } from './VmRunningView.js'

// overridePredicate links a VmV86 to a per-asset override, which takes
// precedence over the V86Image's asset when the VM boots. Mirrors sdk/vm
// PredV86{Kernel,Rootfs,Bios,Wasm}Override, bare IRIs like PRED_V86_IMAGE.
const overridePredicate: Record<OverrideSlot, string> = {
  kernel: 'v86/kernel-override',
  rootfs: 'v86/rootfs-override',
  bios: 'v86/bios-override',
  wasm: 'v86/wasm-override',
}

// VmEdges are the objects a VM boots from.
interface VmEdges {
  imageKey: string
  // imageLabel names the linked image for display, or is "" when it has no
  // readable metadata.
  imageLabel: string
  overrides: Record<OverrideSlot, string>
}

// readVmV86 reads the VmV86 block of an object, or null when it has none.
async function readVmV86(
  objectState: IObjectState,
  signal: AbortSignal,
): Promise<VmV86 | null> {
  using cursor = await objectState.accessWorldState(undefined, signal)
  const blockResp = await cursor.getBlock({}, signal)
  if (!blockResp.found || !blockResp.data) return null
  return VmV86.fromBinary(blockResp.data)
}

// watchVmV86 yields the VmV86 at objectKey and again at each new revision.
async function* watchVmV86(
  world: IWorldState,
  objectKey: string,
  signal: AbortSignal,
): AsyncIterable<VmV86> {
  if (!objectKey) return
  const objectState = await world.getObject(objectKey, signal)
  if (!objectState) return
  using object = objectState
  let revision = (await object.getRootRef(signal)).rev ?? 0n
  for (;;) {
    const vm = await readVmV86(object, signal)
    if (!vm) return
    yield vm
    revision =
      (await object.waitRev(revision + 1n, false, signal)).rev ?? revision
  }
}

// readVmEdges reads the image and asset override edges of the VM at
// objectKey. A missing edge reads as "".
async function readVmEdges(
  world: IWorldState,
  objectKey: string,
  signal: AbortSignal,
): Promise<VmEdges> {
  const subject = keyToIRI(objectKey)
  const lookup = async (predicate: string) => {
    const resp = await world.lookupGraphQuads(
      subject,
      predicate,
      undefined,
      undefined,
      1,
      signal,
    )
    const obj = resp.quads?.[0]?.obj
    return obj ? iriToKey(obj) : ''
  }
  const [imageKey, kernel, rootfs, bios, wasm] = await Promise.all([
    lookup(PRED_V86_IMAGE),
    lookup(overridePredicate.kernel),
    lookup(overridePredicate.rootfs),
    lookup(overridePredicate.bios),
    lookup(overridePredicate.wasm),
  ])
  const image = imageKey
    ? await readV86Image(world, imageKey, signal)
    : undefined
  return {
    imageKey,
    imageLabel: image ? formatV86ImageLabel(image.image) : '',
    overrides: { kernel, rootfs, bios, wasm },
  }
}

// watchVmEdges yields the edges of the VM at objectKey and again after each
// world change.
async function* watchVmEdges(
  world: IWorldState,
  objectKey: string,
  signal: AbortSignal,
): AsyncIterable<VmEdges> {
  let seqno = (await world.getSeqno(signal)).seqno
  for (;;) {
    yield await readVmEdges(world, objectKey, signal)
    seqno = (await world.waitSeqno(seqno + 1n, signal)).seqno
  }
}

// requestVmState asks the VM at objectKey for a target state. A start first
// approves the VM's process binding, which has the Space run the VM; the
// approval outlasts a stop, so the runner commits it. A start needs contents
// to approve the binding.
async function requestVmState(
  world: IWorldState,
  contents: SpaceContents | null,
  objectKey: string,
  target: VmState,
): Promise<void> {
  if (target === VmState.VmState_RUNNING) {
    await contents?.setProcessBinding(objectKey, VmV86TypeID, true)
  }
  const op = SetV86StateOp.create({ objectKey, state: target })
  await world.applyWorldOp(SET_V86_STATE_OP_ID, SetV86StateOp.toBinary(op), '')
}

// writeOverride moves the override edge of one slot from prevKey to nextKey,
// where "" means no override.
async function writeOverride(
  world: IWorldState,
  objectKey: string,
  slot: OverrideSlot,
  prevKey: string,
  nextKey: string,
): Promise<void> {
  const subject = keyToIRI(objectKey)
  const predicate = overridePredicate[slot]
  if (prevKey) {
    await world.deleteGraphQuad(subject, predicate, keyToIRI(prevKey))
  }
  if (nextKey) {
    await world.setGraphQuad(subject, predicate, keyToIRI(nextKey))
  }
}

// useVmV86 watches the VM at objectKey and reads what it boots from, with
// defaults for every field not yet loaded.
function useVmV86(
  worldState: ObjectViewerComponentProps['worldState'],
  objectKey: string,
) {
  // Watch the VM.
  const vmResource = useStreamingResource(
    worldState,
    useCallback(
      (world: IWorldState, signal: AbortSignal) =>
        watchVmV86(world, objectKey, signal),
      [objectKey],
    ),
    [],
  )

  // Watch its edges, and read the UnixFS objects an override may name.
  const edgesResource = useStreamingResource(
    worldState,
    useCallback(
      (world: IWorldState, signal: AbortSignal) =>
        watchVmEdges(world, objectKey, signal),
      [objectKey],
    ),
    [],
  )
  const unixfsResource = useResource(
    worldState,
    (world: IWorldState, signal: AbortSignal) =>
      listObjectsWithType(world, UnixFSTypeID, signal),
    [],
  )

  const vm = vmResource.value ?? VmV86.create()
  const edges = edgesResource.value
  return {
    loaded: !!vmResource.value,
    loadError: vmResource.error,
    name: vm.name || objectKey,
    state: vm.observedState ?? VmState.VmState_STOPPED,
    desired: vm.state ?? VmState.VmState_STOPPED,
    errorMessage: vm.errorMessage?.trim() ?? '',
    memoryMb: vm.config?.memoryMb ?? 0,
    mounts: vm.config?.mounts ?? [],
    imageKey: edges?.imageKey ?? '',
    imageLabel: edges?.imageLabel ?? '',
    overrides: edges?.overrides ?? null,
    unixfsKeys: unixfsResource.value ?? [],
  }
}

// useRequest runs one action at a time from a view and keeps the failure of
// the last one until the next starts.
function useRequest(): [string, (action: () => Promise<void>) => void] {
  const [error, setError] = useState('')
  const request = useCallback((action: () => Promise<void>) => {
    setError('')
    action().catch((err: unknown) => {
      setError(err instanceof Error ? err.message : String(err))
    })
  }, [])
  return [error, request]
}

// VmV86Viewer displays a V86 virtual machine. A stopped or failed VM shows
// its machine card with Start or Reset; a running VM shows its screen and
// serial console.
export default function VmV86Viewer({
  objectInfo,
  worldState,
  spaceContents,
}: ObjectViewerComponentProps) {
  const objectKey = getObjectKey(objectInfo)
  const contextContentsResource = SpaceContentsContext.useContext()
  const contents = useResourceValue(spaceContents ?? contextContentsResource)
  const world = worldState.value
  const vm = useVmV86(worldState, objectKey)
  const [operationError, request] = useRequest()

  // setVmState requests a target state from the view.
  const setVmState = (target: VmState) => {
    if (!world) return
    request(() => requestVmState(world, contents, objectKey, target))
  }

  // Hold the view until the VM record loads, so no default stands in for it.
  if (!vm.loaded) {
    return (
      <div className="bg-background-primary flex h-full w-full items-center justify-center p-4">
        {vm.loadError ? (
          <VmErrorBanner
            title="Could not load this virtual machine"
            message={vm.loadError.message}
          />
        ) : (
          <LoadingCard
            view={{
              state: 'active',
              title: 'Loading virtual machine',
              detail: 'Reading its settings from this Space.',
            }}
            className="w-full max-w-sm"
          />
        )}
      </div>
    )
  }

  const running =
    vm.state !== VmState.VmState_ERROR &&
    (vm.desired === VmState.VmState_RUNNING ||
      vm.state !== VmState.VmState_STOPPED)
  return (
    <div className="bg-background-primary flex h-full w-full flex-col overflow-hidden">
      {running ? (
        <VmRunningView
          objectKey={objectKey}
          name={vm.name}
          state={vm.state}
          stopRequested={vm.desired === VmState.VmState_STOPPED}
          onStop={() => setVmState(VmState.VmState_STOPPED)}
          operationError={operationError}
        />
      ) : (
        <>
          <VmHeader name={vm.name} state={vm.state} />
          <VmMachineCard
            {...vm}
            operationError={operationError}
            startDisabled={!world || !contents}
            onStart={() => setVmState(VmState.VmState_RUNNING)}
            onReset={() => setVmState(VmState.VmState_STOPPED)}
            onOverrideChange={(slot, prevKey, nextKey) =>
              world &&
              request(() =>
                writeOverride(world, objectKey, slot, prevKey, nextKey),
              )
            }
          />
        </>
      )}
    </div>
  )
}
