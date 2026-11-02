#!/bin/sh
# Ookla's licence forbids redistributing its CLI, so it is fetched from Ookla on first run.
set -eu

VERSION="${SPEEDTEST_VERSION:-1.2.0}"
BIN=/opt/speedtest/speedtest

if [ ! -x "$BIN" ]; then
  case "$(uname -m)" in
    aarch64)        ARCH=aarch64 ;;
    armv7l|armv6l)  ARCH=armhf ;;
    x86_64)         ARCH=x86_64 ;;
    i?86)           ARCH=i386 ;;
    *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
  esac
  mkdir -p /opt/speedtest
  wget -qO- "https://install.speedtest.net/app/cli/ookla-speedtest-${VERSION}-linux-${ARCH}.tgz" \
    | tar xz -C /opt/speedtest speedtest
fi

exec "$BIN" --accept-license --accept-gdpr "$@"
