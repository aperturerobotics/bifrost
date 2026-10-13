# WasiVM

WasiVM runs Linux programs compiled to WebAssembly without a Linux kernel. Its
kernel is a Go package whose host functions implement the system calls and
WASI imports those programs make: processes, signals, pipes, a terminal and a
filesystem over a root directory. A BusyBox root filesystem built for
wasm32-wasi boots to an interactive shell.

WasiVM is experimental. Spacewave does not include it unless a build opts in.
A Space that loads the `spacewave-wasivm` plugin can create WebAssembly VMs:
each one boots the kernel on a Space filesystem or a published WasiVM image
and serves its console. A VM on a published image reads its root filesystem
from the CDN as it runs, so creating one copies nothing.

## Layout

| Path | Contents |
| --- | --- |
| `controller.go` | The plugin: it registers the `vm/wasivm` ObjectType, its operations, viewer and wizard, and the `vm/image/wasivm` block type. |
| `handler.go` | The ObjectType and operation handlers the core plugin calls. |
| `vm.go` | The VM resource, which runs a kernel for each run of a VM. |
| `rootfs.go` | Finds the root filesystem of a VM: its own, its image's, or the published image's on the CDN. |
| `console.go` | The console of each running VM, shared by every attached viewer. |
| `*.tsx` | The VM viewer and the wizard that creates a VM. |
| `wasivm-images.ts` | Lists the images published on the CDN and links one into a Space. |
| `kernel/` | The kernel: processes, the filesystem, signals, the terminal and the WASI imports. |
| `kernel/engine/` | The interface the kernel runs guests through. |
| `kernel/engine/native/` | Its implementation on [wazero](https://wazero.io). |
| `cmd/wasivm/` | A command that runs one program under the kernel, for debugging it. |
| `rootfs/` | The script that builds the BusyBox root filesystem. |

## How the kernel works

Each process is a guest module instance. The guest calls the kernel through
`wasi_snapshot_preview1` for files, clocks and arguments, and through `env`
imports for the rest of Linux: `vfork` and `execve`, `wait4`, `kill` and
`sigaction`, terminal ioctls, `ppoll` and more.

- `vfork` and `setjmp` capture the guest's execution state with the engine's
  snapshots and resume from them, restoring the C stack pointer.
- Before a guest runs, the kernel exports its stack pointer and adds a signal
  trampoline that calls a handler from the guest's function table.
- A process that blocks in a read, a wait or a sleep waits on a channel, so a
  signal interrupts it at once.
- The filesystem holds `/dev`, `/proc` and `/tmp` in memory over the root
  directory, which it reads as the guest looks paths up.

## Building the root filesystem

The root filesystem comes from the SkiffOS `virt/wasm` configuration. Building
it takes the host tools Buildroot needs.

```sh
plugin/wasivm/rootfs/build-rootfs.sh [skiffos-dir]
```

Without a directory, the script clones SkiffOS into `.tmp/SkiffOS`. It writes
the tree to `plugin/wasivm/rootfs/out`, which Git ignores.

## Running a program

```sh
go run ./plugin/wasivm/cmd/wasivm --root plugin/wasivm/rootfs/out
go run ./plugin/wasivm/cmd/wasivm -- busybox.wasm sh -c 'echo hello'
```

The first boots the root filesystem's `/bin/sh`; `--init` names another
program. The second runs one BusyBox binary, which also provides its applets.
Flags anywhere on the line belong to `wasivm`, so `--` passes the rest to the
program. `--debug` logs every system call.

## Testing

```sh
go test -timeout=300s ./plugin/wasivm/...
```

The tests that boot BusyBox skip when `rootfs/out` is missing. The tests that
compare WasiVM with Linux in Docker run with `-tags docker_syscall_tests`.
