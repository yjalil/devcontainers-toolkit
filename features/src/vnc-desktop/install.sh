#!/usr/bin/env bash
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

apt-get update -y
apt-get install -y --no-install-recommends \
  # Electron / Chromium shared library dependencies
  libglib2.0-0 libgbm1 libnss3 libatk1.0-0 libatk-bridge2.0-0 \
  libcups2 libdrm2 libxkbcommon0 libxcomposite1 libxdamage1 \
  libxfixes3 libxrandr2 libpango-1.0-0 libcairo2 libasound2 libatspi2.0-0 \
  # VNC server + window manager + web client
  tigervnc-standalone-server tigervnc-common \
  fluxbox \
  novnc websockify

rm -rf /var/lib/apt/lists/*

# Install the VNC start helper into a known location on PATH.
install -m 0755 "$(dirname "$0")/start-vnc.sh" /usr/local/bin/start-vnc

echo "vnc-desktop feature: installed (run 'start-vnc' to launch; ports VNC=${VNCPORT:-5901} noVNC=${NOVNCPORT:-6080})"