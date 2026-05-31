#!/usr/bin/env bash
# Root-level system libraries for the Haskell toolchain (GHC) via mise.
# GHC/cabal/stack come from the project's .mise.toml, NOT from here.
# build-essential, git, gnupg are provided by the base image.
# Sources: haskellstack.org install docs + ghcup system-deps issue + production Haskell images.
set -e
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y --no-install-recommends \
  libffi-dev \
  libgmp-dev \
  libtinfo-dev \
  libncurses-dev \
  zlib1g-dev \
  xz-utils
rm -rf /var/lib/apt/lists/*
echo "haskell feature: build dependencies installed (declare ghc in .mise.toml)"