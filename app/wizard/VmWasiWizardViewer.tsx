import { useCallback, useState } from 'react'
import { LuTerminal } from 'react-icons/lu'

import { useResource } from '@aptre/bldr-sdk/hooks/useResource.js'
import { UnixFSTypeID } from '@s4wave/sdk/unixfs/type.js'
import {
  CREATE_WASIVM_OP_ID,
  DEFAULT_WASIVM_INIT_PATH,
} from '@s4wave/sdk/vm/wasivm.js'
import { CreateWasiVmOp } from '@s4wave/sdk/vm/wasivm.pb.js'
import { listObjectsWithType } from '@s4wave/sdk/world/types/types.js'
import type { IWorldState } from '@s4wave/sdk/world/world-state.js'
import type { ObjectViewerComponentProps } from '@s4wave/web/object/object.js'
import { LoadingCard } from '@s4wave/web/ui/loading/LoadingCard.js'
import { toast } from '@s4wave/web/ui/toaster.js'

import { buildObjectKey } from '../space/create-op-builders.js'
import { useWizardState } from './useWizardState.js'
import { WizardField } from './WizardField.js'
import { WizardFieldFrame } from './WizardFieldFrame.js'
import { wizardInputClassName } from './wizard-field-styles.js'
import { WizardShell } from './WizardShell.js'

// VmWasiWizardViewer creates a WasiVM: its name, the UnixFS object holding
// its root filesystem, and the init program it runs.
export function VmWasiWizardViewer(props: ObjectViewerComponentProps) {
  const ws = useWizardState(props, undefined)
  const { state } = ws
  const [rootfsChoice, setRootfsChoice] = useState('')
  const [initPath, setInitPath] = useState('')

  // List the filesystems the VM can run on; the first is the default.
  const unixfsKeys =
    useResource(
      props.worldState,
      (world: IWorldState, signal: AbortSignal) =>
        listObjectsWithType(world, UnixFSTypeID, signal),
      [],
    ).value ?? []
  const rootfsKey = rootfsChoice || unixfsKeys[0] || ''
  const name = ws.localName.trim()

  // Create the VM, then replace the wizard with it.
  const handleFinalize = useCallback(async () => {
    if (!state || ws.creating || !name || !rootfsKey) return
    ws.setCreating(true)
    try {
      await ws.persistDraftState()
      const objectKey = await buildObjectKey(ws.spaceWorld, 'vm/wasivm/', name)
      const op = CreateWasiVmOp.create({
        objectKey,
        name,
        rootfsObjectKey: rootfsKey,
        config: { initPath: initPath.trim() },
        timestamp: new Date(),
      })
      await ws.spaceWorld.applyWorldOp(
        CREATE_WASIVM_OP_ID,
        CreateWasiVmOp.toBinary(op),
        ws.sessionPeerId,
      )
      await ws.spaceWorld.deleteObject(ws.objectKey)
      toast.success(`Created ${name}`)
      ws.navigateToObjects([objectKey])
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to create VM')
    } finally {
      ws.setCreating(false)
    }
  }, [state, ws, name, rootfsKey, initPath])

  if (!state) {
    return (
      <div className="flex flex-1 items-center justify-center p-4">
        <LoadingCard
          view={{
            state: 'active',
            title: 'Loading wizard',
            detail: 'Preparing the virtual machine settings.',
          }}
          className="w-full max-w-sm"
        />
      </div>
    )
  }

  return (
    <WizardShell
      title={
        <>
          <LuTerminal className="mr-2 size-4 shrink-0" />
          New WebAssembly VM
        </>
      }
      step={state.step ?? 0}
      localName={ws.localName}
      onUpdateName={ws.handleUpdateName}
      onBack={() => void ws.handleBack()}
      onCancel={() => void ws.handleCancel()}
      nameLabel="VM name"
      creating={ws.creating}
      onFinalize={() => void handleFinalize()}
      canFinalize={!!name && !!rootfsKey}
    >
      <WizardFieldFrame
        label="Root filesystem"
        help={
          unixfsKeys.length === 0
            ? 'Add a filesystem to this Space first; the VM runs its programs.'
            : 'The VM runs on a copy of this filesystem; its writes are discarded when it stops.'
        }
      >
        <select
          value={rootfsKey}
          disabled={unixfsKeys.length === 0}
          onChange={(e) => setRootfsChoice(e.target.value)}
          className={`${wizardInputClassName} w-full rounded-md border px-2 font-mono`}
        >
          {unixfsKeys.map((key) => (
            <option key={key} value={key}>
              {key}
            </option>
          ))}
        </select>
      </WizardFieldFrame>
      <WizardField
        label="Init program"
        help="The WebAssembly program the VM starts first, as a path in its filesystem."
        variant="compactMono"
        placeholder={DEFAULT_WASIVM_INIT_PATH}
        value={initPath}
        onChange={(e) => setInitPath(e.target.value)}
      />
    </WizardShell>
  )
}
