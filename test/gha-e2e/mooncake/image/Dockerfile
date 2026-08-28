# The Mooncake image used by the e2e case.
#
# It is built in-repo rather than pulled from an external registry, for the same
# reason as test/gha-e2e/jindo/oss-emulator: e2e runs on every PR, an external
# image going away turns the whole pipeline red, and the two scripts the image
# provides for Fluid to invoke have to be reviewable.
#
# The only external dependency is Mooncake's official package on PyPI.
# python:3.12.13-slim — slim is required: the full variant carries an extra
# ~700MB build toolchain, growing the image from ~470MB to ~1.9GB, and e2e also
# has to kind-load it onto the node and keep a second copy there.
FROM python:3.12.13-slim@sha256:229a2c5bfa27522db7815ea81f9bed70af17ccb9de9fc7ad142b1877b5830d36

RUN apt-get update && apt-get install -y --no-install-recommends \
        libcurl4 \
        libibverbs1 \
        rdma-core \
        librdmacm1 \
        libnuma1 \
        liburing2 \
        curl \
        jq \
    && rm -rf /var/lib/apt/lists/*

# Install from wheels only (--only-binary) so no package build or setup script
# runs at install time, and pin every resolved version explicitly so the image
# e2e builds stays the same from run to run.
RUN pip install --no-cache-dir --only-binary=:all: \
        mooncake-transfer-engine-non-cuda==0.3.12.post1 \
        nvidia-cuda-runtime-cu12==12.8.90

ENV LD_LIBRARY_PATH=/usr/local/lib/python3.12/site-packages/nvidia/cuda_runtime/lib:$LD_LIBRARY_PATH

# 50051 master RPC / 50052 worker data / 8080 metadata service / 9003 master metrics / 9300 worker http
EXPOSE 50051 50052 8080 9003 9300

# The two scripts Fluid invokes:
#   custom-entrypoint.sh  starts the right process based on the role (master/worker)
#   reportSummary.sh      collects cache usage and emits it as JSON in the format Fluid expects
COPY custom-entrypoint.sh /custom-entrypoint.sh
COPY reportSummary.sh /reportSummary.sh
RUN chmod +x /custom-entrypoint.sh /reportSummary.sh

# Run the components as an unprivileged user. Every listening port is above
# 1024, the components write nothing to the container filesystem, and the tiered
# store quota is handed to Mooncake as a process memory segment rather than a
# directory, so no root capability is required at runtime.
RUN useradd --create-home --uid 10001 --user-group mooncake
USER 10001

CMD ["python3"]
