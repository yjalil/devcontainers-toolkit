#!/usr/bin/env bash
# Root-level system libraries for compiling Ruby via mise (ruby-build).
# Ruby itself comes from the project's .mise.toml, NOT from here.
# build-essential is provided by the base image; this adds only the extra headers.
# Sources: official Rails install guide + rbenv/ruby-build wiki.
set -e
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y --no-install-recommends \
  libssl-dev \
  libyaml-dev \
  zlib1g-dev \
  libgmp-dev \
  libffi-dev \
  libreadline-dev \
  libncurses-dev
rm -rf /var/lib/apt/lists/*
echo "ruby feature: build dependencies installed (declare ruby in .mise.toml)"