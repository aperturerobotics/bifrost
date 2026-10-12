import type { ClientResourceRef } from '@aptre/bldr-sdk/resource/client.js'
import { Resource } from '@aptre/bldr-sdk/resource/resource.js'
import type { MessageStream } from 'starpc'

import type { TerminalFrame } from '../terminal/terminal.pb.js'
import { WasiVmConsoleServiceClient } from './wasivm_srpc.pb.js'

// WasiVmTypeID is the object type identifier for WasiVm objects. Mirrors
// WasiVmTypeID in sdk/vm/wasivm.go.
export const WasiVmTypeID = 'vm/wasivm'

// CREATE_WASIVM_OP_ID is the operation type ID for CreateWasiVmOp.
export const CREATE_WASIVM_OP_ID = 'vm/wasivm/create'

// SET_WASIVM_STATE_OP_ID is the operation type ID for SetWasiVmStateOp.
export const SET_WASIVM_STATE_OP_ID = 'vm/wasivm/set-state'

// PRED_WASIVM_ROOTFS is the graph predicate linking a WasiVm to the UnixFS
// object holding its root filesystem. Mirrors PredWasiVmRootfs in
// sdk/vm/wasivm.go.
export const PRED_WASIVM_ROOTFS = 'wasivm/rootfs'

// DEFAULT_WASIVM_INIT_PATH is the init program a WasiVm runs when its config
// names none. Mirrors DefaultWasiVmInitPath in sdk/vm/wasivm.go.
export const DEFAULT_WASIVM_INIT_PATH = '/bin/sh'

// WasiVmHandle is the typed resource of a WasiVm object.
export class WasiVmHandle extends Resource {
  private console: WasiVmConsoleServiceClient

  constructor(resourceRef: ClientResourceRef) {
    super(resourceRef)
    this.console = new WasiVmConsoleServiceClient(resourceRef.client)
  }

  // connectConsole attaches to the console of the running VM. The stream
  // waits for the VM to start, replays its recent output, and ends with EXIT
  // when the VM stops.
  public connectConsole(
    frames: MessageStream<TerminalFrame>,
    abortSignal?: AbortSignal,
  ): MessageStream<TerminalFrame> {
    return this.console.Console(frames, abortSignal)
  }
}
