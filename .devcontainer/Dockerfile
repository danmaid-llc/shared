FROM mcr.microsoft.com/devcontainers/go:2-1.27-trixie

RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    picocom \
    usbutils \
    && apt-get clean -y && rm -rf /var/lib/apt/lists/*