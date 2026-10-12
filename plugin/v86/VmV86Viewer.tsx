import { useCallback, useState } from 'react'

import type { ObjectViewerComponentProps } from '@s4wave/web/object/object.js'
import { getObjectKey } from '@s4wave/web/object/object.js'
import {
  useResource,
  useResourceValue,
} from '@aptre/bldr-sdk/hooks/useResource.js'
import { useStreamingResource } from '@aptre/bldr-sdk/hooks/useStreamingResource.js'
import { SpaceContentsContext } from '@s4wave/web/contexts/contexts.js'
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

import { VmHeader } from './VmHeader.js'
import { VmMachineCard, type OverrideSlot } from './VmMachineCard.js'
import { VmRunningView } from './VmRunningView.js'

// overridePredicate links a VmV86 to a per-asset override, which takes
// precedence over the V86Image's asset when the VM boots. Mirrors sdk/vm
// PredV86{Kernel,Rootfs,Bios,Wasm}Override.
const overridePredicate: Record<OverrideSlot, string> = {
  kernel: '<v86/kernel-override>',
  rootfs: '<v86/rootfs-override>',
  bios: '<v86/bios-override>',
  wasm: '<v86/wasm-override>',
}

// VmEdges are the objects a VM boots from.
interface VmEdges {
  imageKey: string
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
  return { imageKey, overrides: { kernel, rootfs, bios, wasm } }
}

// requestVmState asks the VM at objectKey for a target state. The Space binds
// the VM's process before a start and unbinds it after a stop; a start needs
// contents to bind it.
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
  if (target === VmState.VmState_STOPPED) {
    await contents?.setProcessBinding(objectKey, VmV86TypeID, false)
  }
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

  // Read its edges and the UnixFS objects an override may name.
  const edgesResource = useResource(
    worldState,
    (world: IWorldState, signal: AbortSignal) =>
      readVmEdges(world, objectKey, signal),
    [objectKey],
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
    name: vm.name || objectKey,
    state: vm.observedState ?? VmState.VmState_STOPPED,
    desired: vm.state ?? VmState.VmState_STOPPED,
    errorMessage: vm.errorMessage?.trim() ?? '',
    memoryMb: vm.config?.memoryMb ?? 0,
    mounts: vm.config?.mounts ?? [],
    imageKey: edges?.imageKey ?? '',
    overrides: edges?.overrides ?? null,
    retryEdges: edgesResource.retry,
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
              request(async () => {
                await writeOverride(world, objectKey, slot, prevKey, nextKey)
                vm.retryEdges()
              })
            }
          />
        </>
      )}
    </div>
  )
}
