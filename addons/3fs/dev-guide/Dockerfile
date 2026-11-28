# Base image
FROM ubuntu@sha256:09506232a8004baa32c47d68f1e5c307d648fdd59f5e7eaa42aaf87914100db3

# Arguments
ARG FOUNDATIONDB_TAG=7.3.59
ARG FOUNDATIONDB_VERSION=${FOUNDATIONDB_TAG}-1
ARG LIBFUSE_TAG=fuse-3.16.1
ARG LIBFUSE_VERSION=3.16.1
ARG RUSTUP_VERSION=1.27.1
ARG RUSTUP_INIT_SHA256=6aeece6993e902708983b209d04c0d1dbb14ebb405ddb87def578d41f920f56d

# Install system dependencies and build tools
RUN apt update && \
    apt install -y \
        cmake libuv1-dev liblz4-dev liblzma-dev libdouble-conversion-dev \
        libprocps-dev libdwarf-dev libunwind-dev libaio-dev libgflags-dev \
        libgoogle-glog-dev libgtest-dev libgmock-dev clang-format-14 clang-14 \
        clang-tidy-14 lld-14 libgoogle-perftools-dev google-perftools libssl-dev \
        ccache gcc-12 g++-12 libboost-all-dev git meson ninja-build lsb-release ca-certificates curl && \
    curl -fsSLO --proto '=https' --tlsv1.2 https://apache.jfrog.io/artifactory/arrow/$(lsb_release --id --short | tr 'A-Z' 'a-z')/apache-arrow-apt-source-latest-$(lsb_release --codename --short).deb && \
    apt install -y -V ./apache-arrow-apt-source-latest-$(lsb_release --codename --short).deb && \
    apt update && \
    apt install -y -V libarrow-dev && \
    rm -rf /var/lib/apt/lists/* apache-arrow-apt-source-latest-$(lsb_release --codename --short).deb

# Install Rust
RUN curl -fsSL --proto '=https' --tlsv1.2 -o rustup-init "https://static.rust-lang.org/rustup/archive/${RUSTUP_VERSION}/x86_64-unknown-linux-gnu/rustup-init" && \
    echo "${RUSTUP_INIT_SHA256} *rustup-init" | sha256sum -c - && \
    chmod +x rustup-init && \
    ./rustup-init -y --no-modify-path --default-toolchain 1.90.0 && \
    rm rustup-init

# Install FoundationDB client
RUN curl -fsSLO --proto '=https' --tlsv1.2 https://github.com/apple/foundationdb/releases/download/${FOUNDATIONDB_TAG}/foundationdb-clients_${FOUNDATIONDB_VERSION}_amd64.deb && \
    dpkg -i ./foundationdb-clients_${FOUNDATIONDB_VERSION}_amd64.deb && \
    rm -f foundationdb-clients_${FOUNDATIONDB_VERSION}_amd64.deb

# Build and install libfuse
RUN curl -fsSLO --proto '=https' --tlsv1.2 https://github.com/libfuse/libfuse/releases/download/${LIBFUSE_TAG}/fuse-${LIBFUSE_VERSION}.tar.gz && \
    tar -zxvf fuse-${LIBFUSE_VERSION}.tar.gz && \
    cd fuse-${LIBFUSE_VERSION} && \
    mkdir build && \
    cd build && \
    meson setup .. && \
    ninja && \
    ninja install && \
    cd ../.. && \
    rm -rf fuse-${LIBFUSE_VERSION} fuse-${LIBFUSE_VERSION}.tar.gz

# Set up environment variables
ENV PATH="/root/.cargo/bin:${PATH}"

WORKDIR /app
