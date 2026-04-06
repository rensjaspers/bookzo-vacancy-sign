FROM golang:1.22-bookworm

ENV DEBIAN_FRONTEND=noninteractive

RUN dpkg --add-architecture armhf && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
      crossbuild-essential-armhf \
      fonts-dejavu-core \
      libc6-dev-armhf-cross \
      libstdc++6-armhf-cross \
      libsdl2-dev \
      libsdl2-dev:armhf \
      libsdl2-ttf-dev \
      libsdl2-ttf-dev:armhf \
      make \
      pkg-config \
      python3 \
      zip && \
    rm -rf /var/lib/apt/lists/*
