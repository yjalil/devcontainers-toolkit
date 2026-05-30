#!/usr/bin/env bash
# Feature install scripts run as root at build time, with the Feature's
# options exposed as uppercased env vars (e.g. INCLUDEDBCLIENTS).
set -euo pipefail

export DEBIAN_FRONTEND=noninteractive

# Always-on investigative core
PACKAGES="jq httpie htop lsof strace ripgrep"

if [ "${INCLUDENETTOOLS:-true}" = "true" ]; then
  PACKAGES="$PACKAGES iputils-ping dnsutils netcat-openbsd net-tools iproute2 traceroute telnet"
fi

if [ "${INCLUDEDBCLIENTS:-true}" = "true" ]; then
  PACKAGES="$PACKAGES postgresql-client redis-tools"
fi

apt-get update -y
# shellcheck disable=SC2086
apt-get install -y --no-install-recommends $PACKAGES
rm -rf /var/lib/apt/lists/*

echo "dev-tools feature: installed [$PACKAGES]"