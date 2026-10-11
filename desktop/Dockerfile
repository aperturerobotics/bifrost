FROM --platform=linux/amd64 debian:stable-slim

ARG ZIG_VERSION=0.16.0

RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl xz-utils cmake ninja-build pkg-config \
      file \
    && rm -rf /var/lib/apt/lists/*

RUN dpkg --add-architecture arm64 \
    && apt-get update && apt-get install -y --no-install-recommends \
         libx11-dev libx11-dev:arm64 \
    && rm -rf /var/lib/apt/lists/*

RUN curl -sL https://ziglang.org/download/${ZIG_VERSION}/zig-x86_64-linux-${ZIG_VERSION}.tar.xz \
    | tar -xJ -C /opt \
    && ln -s /opt/zig-x86_64-linux-${ZIG_VERSION}/zig /usr/local/bin/zig

COPY cross/toolchains/zig-ar cross/toolchains/zig-ranlib /usr/local/bin/
RUN chmod +x /usr/local/bin/zig-ar /usr/local/bin/zig-ranlib

# appimagetool for AppImage packaging at release step 6. The AppImage ships
# with a FUSE-based runtime, but FUSE is not available inside this container;
# invocations set APPIMAGE_EXTRACT_AND_RUN=1 to sidestep that dep. The x86_64
# tool produces both amd64 and arm64 output when called with ARCH=<target>.
# Pinned to the aperturerobotics/appimagetool fork so the release is stable
# and the sha256 stays verifiable (upstream continuous moves under us).
#
# sha256 is verified BEFORE we patch the AppImage type-2 magic bytes. The
# patch zeros three bytes at ELF offset 8 ("AI\x02") that the kernel loader
# ignores but qemu-user's stricter ELF parser rejects as ENOEXEC; without
# it, running this container on a non-amd64 host (Docker Desktop under
# qemu emulation on darwin-arm64) fails with "exec format error". See
# https://github.com/AppImage/AppImageKit/issues/828.
#
# The zeroing uses `dd if=/dev/zero` rather than `printf '\x00'` because
# Debian's /bin/sh (dash) does not interpret `\x` hex escapes; dash would
# emit the literal 9-byte string "\x00\x00\x00" and the magic would stay
# non-zero, still tripping qemu-user on the next RUN.
ARG APPIMAGETOOL_VERSION=20260418
ARG APPIMAGETOOL_X86_64_SHA256=a15012a5f782b93d0e49df93f2ed584e19bb0e0e113d92df060f039c92cb9d94
RUN curl -sL \
      "https://github.com/aperturerobotics/appimagetool/releases/download/${APPIMAGETOOL_VERSION}/appimagetool-x86_64.AppImage" \
      -o /usr/local/bin/appimagetool \
    && echo "${APPIMAGETOOL_X86_64_SHA256}  /usr/local/bin/appimagetool" | sha256sum -c - \
    && dd if=/dev/zero of=/usr/local/bin/appimagetool bs=1 seek=8 count=3 conv=notrunc status=none \
    && chmod +x /usr/local/bin/appimagetool \
    && APPIMAGE_EXTRACT_AND_RUN=1 /usr/local/bin/appimagetool --version

WORKDIR /src
