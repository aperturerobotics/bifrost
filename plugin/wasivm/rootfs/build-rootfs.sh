#!/usr/bin/env bash
# build-rootfs.sh builds the wasm32-wasi BusyBox root filesystem into
# rootfs/out with the SkiffOS virt/wasm configuration.
#
# Usage: plugin/wasivm/rootfs/build-rootfs.sh [skiffos-dir]
#
# Without a directory, it clones SkiffOS into the repository's .tmp/SkiffOS.
# The build needs the host tools Buildroot needs, such as make, gcc, g++,
# python3, wget, cpio, unzip, rsync, bc and file, and Binaryen's wasm-opt.
#
# wasm-opt applies the Asyncify transform to every program so the browser
# engine can suspend it across asynchronous host calls and resume it from a
# snapshot. -g keeps the name section, which names __stack_pointer.
set -euo pipefail

script_dir="$(cd "$(dirname "$0")" && pwd)"
out_dir="$script_dir/out"
repo_dir="$(git -C "$script_dir" rev-parse --show-toplevel)"

# Find or clone SkiffOS.
skiffos_dir="${1:-$repo_dir/.tmp/SkiffOS}"
if [ ! -d "$skiffos_dir" ]; then
	echo "Cloning SkiffOS into $skiffos_dir"
	mkdir -p "$(dirname "$skiffos_dir")"
	git clone --recursive https://github.com/skiffos/SkiffOS.git "$skiffos_dir"
fi
if ! command -v wasm-opt >/dev/null; then
	echo "wasm-opt is required: install Binaryen" >&2
	exit 1
fi
if [ ! -f "$skiffos_dir/Makefile" ]; then
	echo "$skiffos_dir is not a SkiffOS checkout" >&2
	exit 1
fi

# Build the root filesystem.
echo "Building the root filesystem in $skiffos_dir"
(
	cd "$skiffos_dir"
	SKIFF_CONFIG=virt/wasm SKIFF_WORKSPACE=wasi make compile
)

# Replace rootfs/out with the built tree.
rootfs_tar="$skiffos_dir/workspaces/wasi/images/rootfs.tar"
rm -rf "$out_dir"
mkdir -p "$out_dir"
tar -xf "$rootfs_tar" -C "$out_dir"
if [ ! -f "$out_dir/usr/bin/busybox" ]; then
	echo "The root filesystem has no /usr/bin/busybox" >&2
	exit 1
fi

# Apply the Asyncify transform to every WebAssembly program.
find "$out_dir" -type f -print0 | while IFS= read -r -d '' program; do
	if [ "$(head -c 4 "$program" | od -An -tx1 | tr -d ' ')" != "0061736d" ]; then
		continue
	fi
	echo "Applying Asyncify to ${program#"$out_dir"/}"
	wasm-opt -O2 -g --strip-dwarf --asyncify \
		--enable-bulk-memory --enable-sign-ext \
		--enable-mutable-globals --enable-nontrapping-float-to-int \
		"$program" -o "$program.asyncify"
	mv "$program.asyncify" "$program"
done
echo "Built $out_dir"
