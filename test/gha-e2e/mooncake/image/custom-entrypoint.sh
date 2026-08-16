#!/bin/sh
# custom-entrypoint.sh
# Usage: /custom-entrypoint.sh <master|worker|client> <start>

set -e

ROLE="$1"
ACTION="$2"

if [ "$ACTION" != "start" ]; then
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
    # Read the dynamic values from the RuntimeConfig JSON mounted by Fluid
    if [ -z "$FLUID_RUNTIME_CONFIG_PATH" ] || [ ! -f "$FLUID_RUNTIME_CONFIG_PATH" ]; then
      echo "Error: FLUID_RUNTIME_CONFIG_PATH not set or file not found"
      exit 1
    fi

    CONFIG=$(cat "$FLUID_RUNTIME_CONFIG_PATH")

    MASTER_SVC=$(echo "$CONFIG" | jq -r '.master.service.name')
    WORKER_SVC=$(echo "$CONFIG" | jq -r '.worker.service.name')
    QUOTA=$(echo "$CONFIG" | jq -r '.worker.tieredStoreLevels[0].quotas[0] // "1GiB"')

    # Quota format conversion: Fluid hands over K8s-style "1Gi", Mooncake wants "1GB"
    SEGMENT_SIZE=$(echo "$QUOTA" | sed 's/Gi$/GB/; s/Mi$/MB/')

    NAMESPACE="${FLUID_DATASET_NAMESPACE:-default}"
    MASTER_ADDR="${MASTER_SVC}.${NAMESPACE}.svc.cluster.local:50051"
    METADATA_ADDR="http://${MASTER_SVC}.${NAMESPACE}.svc.cluster.local:8080/metadata"
    WORKER_HOST="${POD_NAME}.${WORKER_SVC}.${NAMESPACE}.svc.cluster.local"

    echo "Starting worker: master=$MASTER_ADDR, segment_size=$SEGMENT_SIZE, host=$WORKER_HOST"

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
    echo "Error: client role not yet implemented for Mooncake"
    exit 1
    ;;

  *)
    echo "Error: unknown role '$ROLE'"
    exit 1
    ;;
esac
