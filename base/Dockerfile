FROM debian:stable-slim

# ── System substrate: only packages that (a) need root AND (b) every project wants ──
# Build toolchain, fetch tools, VCS, shell, and minimal in-container usability
# (an editor + ps + pager so a bare container isn't painful). Investigative tooling
# (jq, psql, network debug, etc.) is intentionally NOT here — it lives in the
# dev-tools Feature so this base never drifts into a kitchen sink.
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    pkg-config \
    curl \
    wget \
    git \
    sudo \
    zsh \
    ca-certificates \
    unzip \
    rsync \
    tree \
    lsb-release \
    apt-transport-https \
    gnupg \
    # minimal usability (not "tooling")
    vim \
    nano \
    less \
    procps \
    && rm -rf /var/lib/apt/lists/*

# ── Non-root user (UID/GID 1000, passwordless sudo) ──
RUN groupadd -g 1000 vscode \
    && useradd -m -u 1000 -g 1000 -s /bin/zsh -G sudo vscode \
    && echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers

USER vscode
ENV HOME=/home/vscode

# ── mise: the runtime tool/language manager ──
# This is the ONE userspace tool we bootstrap into the base, because it's the
# *mechanism* every project uses to install its own languages/tools at runtime.
# Languages themselves are NEVER baked here — they come from each project's .mise.toml.
RUN curl https://mise.run | sh

ENV PATH="/home/vscode/.local/share/mise/shims:/home/vscode/.local/bin:$PATH"

WORKDIR /workspaces