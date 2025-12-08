# Base image
FROM ubuntu@sha256:09506232a8004baa32c47d68f1e5c307d648fdd59f5e7eaa42aaf87914100db3

# Arguments
ARG FOUNDATIONDB_TAG=7.3.59
ARG FOUNDATIONDB_VERSION=${FOUNDATIONDB_TAG}-1
ARG LIBFUSE_TAG=fuse-3.16.1
ARG LIBFUSE_VERSION=3.16.1

# Install system dependencies and build tools
RUN apt update && \
    apt install -y \
        cmake libuv1-dev liblz4-dev liblzma-dev libdouble-conversion-dev \
        libprocps-dev libdwarf-dev libunwind-dev libaio-dev libgflags-dev \
        libgoogle-glog-dev libgtest-dev libgmock-dev clang-format-14 clang-14 \
        clang-tidy-14 lld-14 libgoogle-perftools-dev google-perftools libssl-dev \
        ccache gcc-12 g++-12 libboost-all-dev git meson ninja-build lsb-release wget && \
    wget https://apache.jfrog.io/artifactory/arrow/$(lsb_release --id --short | tr 'A-Z''a-z')/apache-arrow-apt-source-latest-$(lsb_release --codename --short).deb && \
    apt install -y -V ./apache-arrow-apt-source-latest-$(lsb_release --codename --short).deb && \
    apt update && \
    apt install -y -V libarrow-dev && \
    rm -rf /var/lib/apt/lists/* apache-arrow-apt-source-latest-$(lsb_release --codename --short).deb

# Install Rust
RUN wget -O rustup-init "https://static.rust-lang.org/rustup/dist/x86_64-unknown-linux-gnu/rustup-init" && \
    echo "f39939f3c83a31eda563840394548455a095842de8467556748c55a9346a4959 *rustup-init" | sha256sum -c - && \
    chmod +x rustup-init && \
    ./rustup-init -y --no-modify-path --default-toolchain 1.90.0 && \
    rm rustup-init

# Install FoundationDB client
RUN wget https://github.com/apple/foundationdb/releases/download/${FOUNDATIONDB_TAG}/foundationdb-clients_${FOUNDATIONDB_VERSION}_amd64.deb && \
    dpkg -i ./foundationdb-clients_${FOUNDATIONDB_VERSION}_amd64.deb && \
    rm -f foundationdb-clients_${FOUNDATIONDB_VERSION}_amd64.deb

# Build and install libfuse
RUN wget https://github.com/libfuse/libfuse/releases/download/${LIBFUSE_TAG}/fuse-${LIBFUSE_VERSION}.tar.gz && \
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
