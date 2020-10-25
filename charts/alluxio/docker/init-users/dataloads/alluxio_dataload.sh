#!/usr/bin/env bash
set -e


function distributedLoad() {
    local path=$1
    local replica=$2
    alluxio fs setReplication --max $replica -R $path
    if [[ $needLoadMetadata == 'true' ]]; then
        alluxio fs distributedLoad -Dalluxio.user.file.metadata.sync.interval=0 --replication $replica $path
    else
        alluxio fs distributedLoad --replication $replica $path
    fi
}

function main() {
    needLoadMetadata="$NEED_LOAD_METADATA"
    if [[ $needLoadMetadata == 'true' ]]; then
        if [[ -d "/data" ]]; then
            du -sh "/data"
        fi
    fi
    paths="$DATA_PATH"
    paths=(${paths//:/ })
    replicas="$PATH_REPLICAS"
    replicas=(${replicas//:/ })
    for((i=0;i<${#paths[@]};i++)) do
        echo -e "distributedLoad on ${path[i]} starts"
        distributedLoad ${paths[i]} ${replicas[i]}
        echo -e "distributedLoad on ${path[i]} ends"
    done
}

main "$@"

