#!/bin/bash
# Component entrypoint invoked by Fluid's CacheRuntime.
# Usage: /custom-entrypoint.sh <master|worker|client> start

set -e

ROLE="$1"
ACTION="$2"

if [[ "$ACTION" != "start" ]]; then
  echo "Error: unsupported action '$ACTION'"
  exit 1
fi

case "$ROLE" in

  master)
    exec mooncake_master \
      -v=1 \
      --rpc_interface=eth0 \
      --enable_http_metadata_server=true \
      --http_metadata_server_host=0.0.0.0 \
      --http_metadata_server_port=8080 \
      --enable_metadata_cleanup_on_timeout=true \
      --client_ttl=10
    ;;

  worker)
    # Fluid mounts runtime.sh alongside the runtime.json named by this variable.
    : "${FLUID_RUNTIME_CONFIG_PATH:?missing runtime config path}"
    source "$(dirname "$FLUID_RUNTIME_CONFIG_PATH")/runtime.sh"
    QUOTA="${WORKER_TIEREDSTORE_LEVEL_0_QUOTA_0:-1GiB}"

    # Fluid reports the quota in Kubernetes units ("1Gi"); Mooncake expects "1GB".
    SEGMENT_SIZE=$(echo "$QUOTA" | sed 's/Gi$/GB/; s/Mi$/MB/')

    NAMESPACE="${FLUID_DATASET_NAMESPACE:-default}"
    MASTER_ADDR="${MASTER_SERVICE_NAME}.${NAMESPACE}.svc.cluster.local:50051"
    METADATA_ADDR="http://${MASTER_SERVICE_NAME}.${NAMESPACE}.svc.cluster.local:8080/metadata"
    WORKER_HOST="${POD_NAME}.${WORKER_SERVICE_NAME}.${NAMESPACE}.svc.cluster.local"

    echo "Starting worker: master=$MASTER_ADDR, segment_size=$SEGMENT_SIZE, host=$WORKER_HOST"

    # The metadata endpoint is plain HTTP and the transfer protocol is TCP
    # because Mooncake exposes no TLS variant for either: the master's built-in
    # metadata server (--enable_http_metadata_server) only speaks HTTP, and the
    # transfer engine only offers "tcp" and "rdma". Both connections stay inside
    # the cluster, addressed by ClusterIP/headless service DNS, and carry cache
    # blocks between components of this runtime only. Put the runtime in a
    # dedicated namespace with a NetworkPolicy, or a service mesh with mTLS, if
    # that traffic needs to be protected on the wire.
    exec mooncake_client \
      --host="$WORKER_HOST" \
      --port=50052 \
      --global_segment_size="$SEGMENT_SIZE" \
      --master_server_address="$MASTER_ADDR" \
      --metadata_server="$METADATA_ADDR" \
      --protocol=tcp \
      --enable_http_server=true \
      --http_port=9300
    ;;

  client)
    # Mooncake is a client-less cache system in Fluid: applications link the
    # Mooncake client library directly, so no client component is deployed.
    echo "Error: client role is not applicable to Mooncake"
    exit 1
    ;;

  *)
    echo "Error: unknown role '$ROLE'"
    exit 1
    ;;
esac
