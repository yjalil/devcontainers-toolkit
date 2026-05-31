#!/usr/bin/env bash
# Root-level system libraries for Lua + LuaRocks via mise.
# Lua itself comes from the project's .mise.toml, NOT from here.
# build-essential is provided by the base image.
set -e
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y --no-install-recommends \
  libreadline-dev \
  unzip
rm -rf /var/lib/apt/lists/*
echo "lua feature: build dependencies installed (declare lua in .mise.toml)"