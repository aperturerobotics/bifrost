import { useCallback, useMemo, useState } from 'react'
import { LuMonitor, LuSquareTerminal } from 'react-icons/lu'

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
  PRED_WASIVM_IMAGE,
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
import { useObjectName } from '@s4wave/web/object/ObjectViewerContext.js'
import { Button } from '@s4wave/web/ui/button.js'
import { LoadingCard } from '@s4wave/web/ui/loading/LoadingCard.js'

import { useVmRequest, watchVmObject } from '@s4wave/app/vm/vm-object.js'
import { VmErrorBanner } from '@s4wave/app/vm/VmErrorBanner.js'
import { VmHeader, VmStopButton } from '@s4wave/app/vm/VmHeader.js'

import { WasiVmDisplay, type WasiVmDisplayConnector } from './WasiVmDisplay.js'
import { WasiVmMachineCard } from './WasiVmMachineCard.js'

// readLinkedKey reads the key of the first object the VM at objectKey links to
// with predicate, or "" when it has none.
async function readLinkedKey(
  world: IWorldState,
  objectKey: string,
  predicate: string,
  signal: AbortSignal,
): Promise<string> {
  const resp = await world.lookupGraphQuads(
    keyToIRI(objectKey),
    predicate,
    undefined,
    undefined,
    1,
    signal,
  )
  const obj = resp.quads?.[0]?.obj
  return obj ? iriToKey(obj) : ''
}

// readRootfsKey reads the key of the object the VM at objectKey takes its
// root filesystem from: its own root filesystem, else its image, else "".
async function readRootfsKey(
  world: IWorldState,
  objectKey: string,
  signal: AbortSignal,
): Promise<string> {
  const rootfs = await readLinkedKey(
    world,
    objectKey,
    PRED_WASIVM_ROOTFS,
    signal,
  )
  return rootfs || readLinkedKey(world, objectKey, PRED_WASIVM_IMAGE, signal)
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
  useObjectName(vm.name)
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

// WasiVmScreenToggleProps configures the switch between console and display.
interface WasiVmScreenToggleProps {
  showDisplay: boolean
  onToggle: () => void
}

// WasiVmScreenToggle switches a running VM's view between its console and
// its display. Its label collapses to the icon in a narrow header.
function WasiVmScreenToggle({
  showDisplay,
  onToggle,
}: WasiVmScreenToggleProps) {
  return (
    <Button
      variant="outline"
      size="toolbarSm"
      aria-label={showDisplay ? 'Show console' : 'Show display'}
      onClick={onToggle}
    >
      {showDisplay ? <LuSquareTerminal /> : <LuMonitor />}
      <span className="hidden @md:inline">
        {showDisplay ? 'Console' : 'Display'}
      </span>
    </Button>
  )
}

// WasiVmScreenProps configures the screen of one run of a VM.
interface WasiVmScreenProps {
  worldState: ObjectViewerComponentProps['worldState']
  objectKey: string
  showDisplay: boolean
}

// WasiVmScreen shows the console or the display of a running VM, each
// attached through the VM's typed resource.
function WasiVmScreen({
  worldState,
  objectKey,
  showDisplay,
}: WasiVmScreenProps) {
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
  const connectDisplay = useMemo<WasiVmDisplayConnector | undefined>(
    () =>
      handle
        ? (input, signal) => handle.connectDisplay(input, signal)
        : undefined,
    [handle],
  )
  return showDisplay ? (
    <WasiVmDisplay connectDisplay={connectDisplay} />
  ) : (
    <TerminalPane
      connectTerminal={connectConsole}
      describeFailure={describeConsoleFailure}
    />
  )
}

// WasiVmViewer displays a WasiVM. A stopped or failed VM shows its machine
// card with Start or Reset; a running VM shows its console or its display.
export default function WasiVmViewer({
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
  const [showDisplay, setShowDisplay] = useState(false)

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
          <WasiVmScreenToggle
            showDisplay={showDisplay}
            onToggle={() => setShowDisplay(!showDisplay)}
          />
        )}
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
        // Each run is a new kernel, so a new run attaches a new screen.
        <WasiVmScreen
          key={vm.runGeneration.toString()}
          worldState={worldState}
          objectKey={objectKey}
          showDisplay={showDisplay}
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
