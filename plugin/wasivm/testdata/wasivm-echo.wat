;; wasivm-echo reads one line from the console and writes it back after
;; "got:". Build with: wat2wasm wasivm-echo.wat -o wasivm-echo.wasm
(module
  (import "wasi_snapshot_preview1" "fd_read" (func $fd_read (param i32 i32 i32 i32) (result i32)))
  (import "wasi_snapshot_preview1" "fd_write" (func $fd_write (param i32 i32 i32 i32) (result i32)))
  (memory (export "memory") 1)
  (data (i32.const 64) "got:")
  (func (export "_start")
    (i32.store (i32.const 32) (i32.const 68))
    (i32.store (i32.const 36) (i32.const 60))
    (drop (call $fd_read (i32.const 0) (i32.const 32) (i32.const 1) (i32.const 48)))
    (i32.store (i32.const 16) (i32.const 64))
    (i32.store (i32.const 20) (i32.add (i32.const 4) (i32.load (i32.const 48))))
    (drop (call $fd_write (i32.const 1) (i32.const 16) (i32.const 1) (i32.const 52)))))
