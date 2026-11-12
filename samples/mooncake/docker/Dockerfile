# Build context for the image used by the Mooncake CacheRuntime sample.
#
# See docs/en/samples/cacheruntime/mooncake_cache_runtime.md
#     docs/zh/samples/cacheruntime/mooncake_cache_runtime.md
#
# Build and push to a registry you control:
#   docker build -t <your-registry>/mooncake:v3 samples/mooncake/docker
#   docker push <your-registry>/mooncake:v3

FROM python:3.12-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
      libcurl4 libibverbs1 rdma-core librdmacm1 libnuma1 liburing2 curl jq \
    && rm -rf /var/lib/apt/lists/*

# Install from wheels only (--only-binary) so no package build/setup script is
# executed at install time, and pin every resolved version explicitly.
RUN pip install --no-cache-dir --only-binary=:all: \
      mooncake-transfer-engine-non-cuda==0.3.12.post1 \
      nvidia-cuda-runtime-cu12==12.8.90

ENV LD_LIBRARY_PATH=/usr/local/lib/python3.12/site-packages/nvidia/cuda_runtime/lib:$LD_LIBRARY_PATH

# master rpc / worker data / metadata / master metrics / worker http
EXPOSE 50051 50052 8080 9003 9300

COPY custom-entrypoint.sh /custom-entrypoint.sh
COPY reportSummary.sh /reportSummary.sh
RUN chmod +x /custom-entrypoint.sh /reportSummary.sh

# Run the components as an unprivileged user. All listening ports are above
# 1024 and the sample writes nothing to the container filesystem, so no root
# capability is required at runtime.
RUN useradd --create-home --uid 10001 --user-group mooncake
USER 10001
