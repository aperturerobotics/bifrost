# bldr callpath

`bldr callpath` prints the static call paths from the functions matching
`--from` to the functions matching `--to`. Use it to answer "can this entry
point reach that call?" across a package tree, for example whether any World
operation can take a lock or read the database it is applying to. The same
command is `spacewave-dev callpath` in an app built on spacewave.

## Usage

```text
bldr callpath --from <regexp> --to <regexp> [flags] <package-pattern>...
```

- `--from` matches the functions paths start from.
- `--to` matches the functions paths end at. A target may lie in any package.
- `--within` matches the package paths a path may pass through. By default a
  path stays inside the modules of the loaded packages.
- `--tests` also loads the packages' test files.
- `--full` prints full package paths. By default the loaded modules' paths are
  trimmed.

Names are SSA names: `pkg.Func`, `(*pkg.T).Method`, `(pkg.T).Method`, and
`pkg.Func$1` for the first closure in `Func`.

## Examples

Every World operation that can take a mutex:

```bash
bldr callpath --from '\)\.ApplyWorldOp$' \
  --to '^\(\*sync\.(RW)?Mutex\)\.R?Lock$' ./core/...
```

Every path from an operation to a read of the store:

```bash
bldr callpath --from '\)\.ApplyWorldOp$' --to '\.sqlDatabase\)\.read$' ./core/worldgraph
```

## Reading the output

Each block is the shortest path from one root to one target:

```text
(*db/world/vlogger.WorldState).ApplyWorldOp
  -> (*db/world/block/tx.WorldState).ApplyWorldOp
  -> (*sync.Mutex).Lock
```

The call graph is a class hierarchy analysis: an interface call reaches every
method of every loaded type that satisfies the interface. A path may therefore
never run, so confirm it in the source. A path that is missing does not exist
in the loaded code, which makes an empty result the useful answer for a "can
never happen" question. A call through a function value reaches every function
of the same signature whose address the loaded code takes.
