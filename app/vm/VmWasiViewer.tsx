import { useCallback, useMemo } from 'react'

import {
  useResource,
  useResourceValue,
} from '@aptre/bldr-sdk/hooks/useResource.js'
import { useStreamingResource } from '@aptre/bldr-sdk/hooks/useStreamingResource.js'
import {
  TerminalPane,
  type TerminalPaneConnector,
} from '@s4wave/app/terminal/TerminalPane.js'
import type { SpaceContents } from '@s4wave/sdk/space/contents.js'
import { VmState } from '@s4wave/sdk/vm/v86.pb.js'
import {
  DEFAULT_WASIVM_INIT_PATH,
  PRED_WASIVM_ROOTFS,
  SET_WASIVM_STATE_OP_ID,
  WasiVmHandle,
  WasiVmTypeID,
} from '@s4wave/sdk/vm/wasivm.js'
import { SetWasiVmStateOp, WasiVm } from '@s4wave/sdk/vm/wasivm.pb.js'
import { iriToKey, keyToIRI } from '@s4wave/sdk/world/graph-utils.js'
import type { IWorldState } from '@s4wave/sdk/world/world-state.js'
import { SpaceContentsContext } from '@s4wave/web/contexts/contexts.js'
import { useAccessTypedHandle } from '@s4wave/web/hooks/useAccessTypedHandle.js'
import type { ObjectViewerComponentProps } from '@s4wave/web/object/object.js'
import { getObjectKey } from '@s4wave/web/object/object.js'
import { LoadingCard } from '@s4wave/web/ui/loading/LoadingCard.js'

import { useVmRequest, watchVmObject } from './vm-object.js'
import { VmErrorBanner } from './VmErrorBanner.js'
import { VmHeader, VmStopButton } from './VmHeader.js'
import { WasiVmMachineCard } from './WasiVmMachineCard.js'

// readRootfsKey reads the key of the root filesystem the VM at objectKey
// runs on, or "" when it has none.
async function readRootfsKey(
  world: IWorldState,
  objectKey: string,
  signal: AbortSignal,
): Promise<string> {
  const resp = await world.lookupGraphQuads(
    keyToIRI(objectKey),
    PRED_WASIVM_ROOTFS,
    undefined,
    undefined,
    1,
    signal,
  )
  const obj = resp.quads?.[0]?.obj
  return obj ? iriToKey(obj) : ''
}

// requestWasiVmState asks the VM at objectKey for a target state. A start
// first approves the VM's process binding, which has the Space run the VM;
// the approval outlasts a stop. A start needs contents to approve the
// binding.
async function requestWasiVmState(
  world: IWorldState,
  contents: SpaceContents | null | undefined,
  objectKey: string,
  target: VmState,
): Promise<void> {
  if (target === VmState.VmState_RUNNING) {
    await contents?.setProcessBinding(objectKey, WasiVmTypeID, true)
  }
  const op = SetWasiVmStateOp.create({ objectKey, state: target })
  await world.applyWorldOp(
    SET_WASIVM_STATE_OP_ID,
    SetWasiVmStateOp.toBinary(op),
    '',
  )
}

// describeConsoleFailure states why the console of a VM could not attach.
function describeConsoleFailure(rawError?: string): string {
  return rawError
    ? `The console disconnected: ${rawError}`
    : 'The console ended before the VM wrote anything. Retry to attach again.'
}

// useWasiVm watches the VM at objectKey and reads its root filesystem, with
// defaults for every field not yet loaded.
function useWasiVm(
  worldState: ObjectViewerComponentProps['worldState'],
  objectKey: string,
) {
  // Watch the VM and read the root filesystem it runs on.
  const vmResource = useStreamingResource(
    worldState,
    useCallback(
      (world: IWorldState, signal: AbortSignal) =>
        watchVmObject(world, objectKey, (b) => WasiVm.fromBinary(b), signal),
      [objectKey],
    ),
    [],
  )
  const rootfsResource = useResource(
    worldState,
    (world: IWorldState, signal: AbortSignal) =>
      readRootfsKey(world, objectKey, signal),
    [objectKey],
  )

  const vm = vmResource.value ?? WasiVm.create()
  const initPath = vm.config?.initPath || DEFAULT_WASIVM_INIT_PATH
  return {
    loaded: !!vmResource.value,
    loadError: vmResource.error,
    name: vm.name || objectKey,
    state: vm.observedState ?? VmState.VmState_STOPPED,
    desired: vm.state ?? VmState.VmState_STOPPED,
    errorMessage: vm.errorMessage?.trim() ?? '',
    runGeneration: vm.runGeneration ?? 0n,
    rootfsKey: rootfsResource.value ?? '',
    init: [initPath, ...(vm.config?.initArgs ?? [])].join(' '),
  }
}

// VmWasiViewer displays a WasiVM. A stopped or failed VM shows its machine
// card with Start or Reset; a running VM shows its console.
export function VmWasiViewer({
  objectInfo,
  worldState,
  spaceContents,
}: ObjectViewerComponentProps) {
  const objectKey = getObjectKey(objectInfo)
  const contextContentsResource = SpaceContentsContext.useContext()
  const contents = useResourceValue(spaceContents ?? contextContentsResource)
  const world = worldState.value
  const vm = useWasiVm(worldState, objectKey)
  const [operationError, request] = useVmRequest()

  // Attach the console through the VM's typed resource.
  const handle = useAccessTypedHandle(
    worldState,
    objectKey,
    WasiVmHandle,
    WasiVmTypeID,
  ).value
  const connectConsole = useMemo<TerminalPaneConnector | undefined>(
    () =>
      handle
        ? (frames, signal) => handle.connectConsole(frames, signal)
        : undefined,
    [handle],
  )

  // setVmState requests a target state from the view.
  const setVmState = (target: VmState) => {
    if (!world) return
    request(() => requestWasiVmState(world, contents, objectKey, target))
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
  const stopRequested = vm.desired === VmState.VmState_STOPPED
  return (
    <div className="bg-background-primary flex h-full w-full flex-col overflow-hidden">
      <VmHeader name={vm.name} state={vm.state}>
        {running && (
          <VmStopButton
            disabled={stopRequested || vm.state === VmState.VmState_STOPPING}
            onStop={() => setVmState(VmState.VmState_STOPPED)}
          />
        )}
      </VmHeader>
      {operationError && running && (
        <div className="shrink-0 p-3">
          <VmErrorBanner title="Operation error" message={operationError} />
        </div>
      )}
      {running ? (
        // Each run is a new kernel, so a new run attaches a new terminal.
        <TerminalPane
          key={vm.runGeneration.toString()}
          connectTerminal={connectConsole}
          describeFailure={describeConsoleFailure}
        />
      ) : (
        <WasiVmMachineCard
          {...vm}
          operationError={operationError}
          startDisabled={!world || !contents}
          onStart={() => setVmState(VmState.VmState_RUNNING)}
          onReset={() => setVmState(VmState.VmState_STOPPED)}
        />
      )}
    </div>
  )
}
