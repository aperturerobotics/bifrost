import type { ClientResourceRef } from '@aptre/bldr-sdk/resource/client.js'
import { Resource } from '@aptre/bldr-sdk/resource/resource.js'
import type { MessageStream } from 'starpc'

import type { TerminalFrame } from '../terminal/terminal.pb.js'
import type { WasiVmDisplayFrame, WasiVmDisplayInput } from './wasivm.pb.js'
import {
  WasiVmConsoleServiceClient,
  WasiVmDisplayServiceClient,
} from './wasivm_srpc.pb.js'

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

// PRED_WASIVM_IMAGE is the graph predicate linking a WasiVm to the
// WasiVmImage it boots when it has no root filesystem of its own. Mirrors
// PredWasiVmImage in sdk/vm/wasivm.go.
export const PRED_WASIVM_IMAGE = 'wasivm/image'

// WasiVmImageTypeID is the object type identifier for WasiVmImage objects.
// Mirrors WasiVmImageTypeID in sdk/vm/wasivm-image.go.
export const WasiVmImageTypeID = 'vm/image/wasivm'

// CREATE_WASIVM_IMAGE_OP_ID is the operation type ID for
// CreateWasiVmImageOp.
export const CREATE_WASIVM_IMAGE_OP_ID = 'vm/image/wasivm/create'

// PRED_WASIVM_IMAGE_ROOTFS is the graph predicate linking a WasiVmImage to
// the UnixFS object holding its root filesystem. Mirrors
// PredWasiVmImageRootfs in sdk/vm/wasivm-image.go.
export const PRED_WASIVM_IMAGE_ROOTFS = 'wasivmimage/rootfs'

// DEFAULT_WASIVM_INIT_PATH is the init program a WasiVm runs when its config
// names none. Mirrors DefaultWasiVmInitPath in sdk/vm/wasivm.go.
export const DEFAULT_WASIVM_INIT_PATH = '/bin/sh'

// WasiVmHandle is the typed resource of a WasiVm object.
export class WasiVmHandle extends Resource {
  private console: WasiVmConsoleServiceClient
  private display: WasiVmDisplayServiceClient

  constructor(resourceRef: ClientResourceRef) {
    super(resourceRef)
    this.console = new WasiVmConsoleServiceClient(resourceRef.client)
    this.display = new WasiVmDisplayServiceClient(resourceRef.client)
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

  // connectDisplay attaches to X display 0 of the running VM. The stream
  // waits for the VM and its X server, sends the whole screen, then each
  // region that changes; input carries the viewer's pointer and keys back.
  public connectDisplay(
    input: MessageStream<WasiVmDisplayInput>,
    abortSignal?: AbortSignal,
  ): MessageStream<WasiVmDisplayFrame> {
    return this.display.Display(input, abortSignal)
  }
}
