#!/bin/bash

testname="jindoruntime secret e2e"

s3_dataset_name="jindo-demo"
s3_dataset_file="test/gha-e2e/jindo/dataset.yaml"
s3_job_name="fluid-test"
s3_job_file="test/gha-e2e/jindo/job.yaml"

multi_oss_dataset_name="jindo-multi-oss-demo"
multi_oss_dataset_template="test/gha-e2e/jindo/multi-oss-dataset.yaml"
multi_oss_job_name="fluid-multi-oss-test"
multi_oss_job_template="test/gha-e2e/jindo/multi-oss-job.yaml"

multi_oss_bucket_a="${JINDO_E2E_OSS_BUCKET_A:-bucketa}"
multi_oss_bucket_b="${JINDO_E2E_OSS_BUCKET_B:-bucketb}"
multi_oss_endpoint_a="${JINDO_E2E_OSS_ENDPOINT_A:-oss.default.svc.cluster.local:9000}"
multi_oss_endpoint_b="${JINDO_E2E_OSS_ENDPOINT_B:-oss.default.svc.cluster.local:9000}"
multi_oss_access_key_a="${JINDO_E2E_OSS_ACCESS_KEY_ID_A:-bucketaadmin}"
multi_oss_secret_key_a="${JINDO_E2E_OSS_ACCESS_KEY_SECRET_A:-bucketasecret}"
multi_oss_access_key_b="${JINDO_E2E_OSS_ACCESS_KEY_ID_B:-bucketbadmin}"
multi_oss_secret_key_b="${JINDO_E2E_OSS_ACCESS_KEY_SECRET_B:-bucketbsecret}"
multi_oss_object_key_a="${JINDO_E2E_OSS_OBJECT_KEY_A:-testfile}"
multi_oss_object_key_b="${JINDO_E2E_OSS_OBJECT_KEY_B:-testfile}"
multi_oss_expected_data_a="${JINDO_E2E_OSS_EXPECTED_DATA_A:-bucket-a-data}"
multi_oss_expected_data_b="${JINDO_E2E_OSS_EXPECTED_DATA_B:-bucket-b-data}"
multi_oss_seed_endpoint=""
multi_oss_backend="${JINDO_E2E_MULTI_OSS_BACKEND:-emulator}"

rendered_dir=""
multi_oss_dataset_file=""
multi_oss_job_file=""

function syslog() {
    echo ">>> $1"
}

function panic() {
    err_msg=$1
    syslog "test \"$testname\" failed: $err_msg"
    exit 1
}

function is_real_oss_mode() {
    [[ -n "${JINDO_E2E_OSS_ENDPOINT:-}" ]]
}

function should_skip_s3_scenario() {
    [[ "${JINDO_E2E_SKIP_S3_SCENARIO:-}" == "1" ]]
}

function should_use_minio_multi_oss() {
    [[ "$multi_oss_backend" == "minio" ]]
}

function should_use_emulator_multi_oss() {
    [[ "$multi_oss_backend" == "emulator" ]]
}

function require_env() {
    var_name=$1
    if [[ -z "${!var_name:-}" ]]; then
        panic "required environment variable $var_name is empty"
    fi
}

function normalize_mount_endpoint() {
    endpoint=$1
    endpoint="${endpoint#http://}"
    endpoint="${endpoint#https://}"
    endpoint="${endpoint%/}"
    echo "$endpoint"
}

function ensure_python_oss2() {
    if python3 - <<'PY' >/dev/null 2>&1
import importlib.util
import sys
sys.exit(0 if importlib.util.find_spec("oss2") else 1)
PY
    then
        return
    fi

    syslog "Installing python module oss2 for real OSS verification"
    python3 -m pip install --user oss2 >/dev/null || panic "failed to install python module oss2"
}

function ensure_python_boto3() {
    if python3 - <<'PY' >/dev/null 2>&1
import importlib.util
import sys
mods = ("boto3", "botocore")
sys.exit(0 if all(importlib.util.find_spec(m) for m in mods) else 1)
PY
    then
        return
    fi

    syslog "Installing python module boto3 for MinIO verification"
    python3 -m pip install --user boto3 >/dev/null || panic "failed to install python module boto3"
}

function seed_real_oss_bucket() {
    bucket_name=$1
    access_key=$2
    secret_key=$3
    bucket_object_key=$4
    file_content=$5

    OSS_BUCKET_NAME="$bucket_name" \
    OSS_ENDPOINT="$multi_oss_seed_endpoint" \
    OSS_ACCESS_KEY="$access_key" \
    OSS_SECRET_KEY="$secret_key" \
    OSS_OBJECT_KEY="$bucket_object_key" \
    OSS_FILE_CONTENT="$file_content" \
    python3 - <<'PY' || exit 1
import os
import sys

import oss2

bucket = oss2.Bucket(
    oss2.Auth(os.environ["OSS_ACCESS_KEY"], os.environ["OSS_SECRET_KEY"]),
    os.environ["OSS_ENDPOINT"],
    os.environ["OSS_BUCKET_NAME"],
)
bucket.put_object(os.environ["OSS_OBJECT_KEY"], os.environ["OSS_FILE_CONTENT"].encode())
value = bucket.get_object(os.environ["OSS_OBJECT_KEY"]).read().decode()
if value != os.environ["OSS_FILE_CONTENT"]:
    raise SystemExit(f"unexpected object content: {value!r}")
print(f"seeded {os.environ['OSS_BUCKET_NAME']}/{os.environ['OSS_OBJECT_KEY']}")
PY
}

function setup_real_oss() {
    require_env JINDO_E2E_OSS_ENDPOINT
    require_env JINDO_E2E_OSS_BUCKET_A
    require_env JINDO_E2E_OSS_BUCKET_B
    require_env JINDO_E2E_OSS_ACCESS_KEY_ID_A
    require_env JINDO_E2E_OSS_ACCESS_KEY_SECRET_A
    require_env JINDO_E2E_OSS_ACCESS_KEY_ID_B
    require_env JINDO_E2E_OSS_ACCESS_KEY_SECRET_B

    ensure_python_oss2

    multi_oss_bucket_a="$JINDO_E2E_OSS_BUCKET_A"
    multi_oss_bucket_b="$JINDO_E2E_OSS_BUCKET_B"
    multi_oss_endpoint_a="$(normalize_mount_endpoint "$JINDO_E2E_OSS_ENDPOINT")"
    multi_oss_endpoint_b="$(normalize_mount_endpoint "$JINDO_E2E_OSS_ENDPOINT")"
    multi_oss_seed_endpoint="${JINDO_E2E_OSS_SEED_ENDPOINT:-$JINDO_E2E_OSS_ENDPOINT}"
    if [[ "$multi_oss_seed_endpoint" != http://* && "$multi_oss_seed_endpoint" != https://* ]]; then
        multi_oss_seed_endpoint="https://$multi_oss_seed_endpoint"
    fi
    multi_oss_access_key_a="$JINDO_E2E_OSS_ACCESS_KEY_ID_A"
    multi_oss_secret_key_a="$JINDO_E2E_OSS_ACCESS_KEY_SECRET_A"
    multi_oss_access_key_b="$JINDO_E2E_OSS_ACCESS_KEY_ID_B"
    multi_oss_secret_key_b="$JINDO_E2E_OSS_ACCESS_KEY_SECRET_B"

    seed_real_oss_bucket "$multi_oss_bucket_a" "$multi_oss_access_key_a" "$multi_oss_secret_key_a" "$multi_oss_object_key_a" "$multi_oss_expected_data_a" || panic "failed to seed real oss bucket $multi_oss_bucket_a"
    seed_real_oss_bucket "$multi_oss_bucket_b" "$multi_oss_access_key_b" "$multi_oss_secret_key_b" "$multi_oss_object_key_b" "$multi_oss_expected_data_b" || panic "failed to seed real oss bucket $multi_oss_bucket_b"
}

function render_multi_oss_files() {
    rendered_dir=$(mktemp -d)
    multi_oss_dataset_file="$rendered_dir/multi-oss-dataset.yaml"
    multi_oss_job_file="$rendered_dir/multi-oss-job.yaml"
    endpoint_a="$(normalize_mount_endpoint "$multi_oss_endpoint_a")"
    endpoint_b="$(normalize_mount_endpoint "$multi_oss_endpoint_b")"

    sed \
        -e "s|__ACCESS_KEY_ID_A__|$multi_oss_access_key_a|g" \
        -e "s|__ACCESS_KEY_SECRET_A__|$multi_oss_secret_key_a|g" \
        -e "s|__ACCESS_KEY_ID_B__|$multi_oss_access_key_b|g" \
        -e "s|__ACCESS_KEY_SECRET_B__|$multi_oss_secret_key_b|g" \
        -e "s|__BUCKET_A__|$multi_oss_bucket_a|g" \
        -e "s|__BUCKET_B__|$multi_oss_bucket_b|g" \
        -e "s|__ENDPOINT_A__|$endpoint_a|g" \
        -e "s|__ENDPOINT_B__|$endpoint_b|g" \
        "$multi_oss_dataset_template" >"$multi_oss_dataset_file"

    sed \
        -e "s|__EXPECTED_DATA_A__|$multi_oss_expected_data_a|g" \
        -e "s|__EXPECTED_DATA_B__|$multi_oss_expected_data_b|g" \
        -e "s|__OBJECT_KEY_A__|$multi_oss_object_key_a|g" \
        -e "s|__OBJECT_KEY_B__|$multi_oss_object_key_b|g" \
        "$multi_oss_job_template" >"$multi_oss_job_file"
}

function seed_minio_bucket() {
    app_name=$1
    access_key=$2
    secret_key=$3
    bucket_name=$4
    object_key=$5
    file_content=$6
    local_port=$7
    port_forward_log="/tmp/${app_name}-port-forward.log"

    ensure_python_boto3

    kubectl port-forward service/${app_name} ${local_port}:9000 >${port_forward_log} 2>&1 &
    pf_pid=$!

    for i in $(seq 1 20); do
        if python3 - <<PY >/dev/null 2>&1
import socket
s = socket.socket()
s.settimeout(1)
try:
    s.connect(("127.0.0.1", ${local_port}))
finally:
    s.close()
PY
        then
            break
        fi
        if ! kill -0 ${pf_pid} >/dev/null 2>&1; then
            cat ${port_forward_log} >&2 || true
            panic "port-forward for ${app_name} exited unexpectedly"
        fi
        sleep 1
    done

    MINIO_ENDPOINT="http://127.0.0.1:${local_port}" \
    MINIO_ACCESS_KEY="${access_key}" \
    MINIO_SECRET_KEY="${secret_key}" \
    MINIO_BUCKET_NAME="${bucket_name}" \
    MINIO_OBJECT_KEY="${object_key}" \
    MINIO_FILE_CONTENT="${file_content}" \
    python3 - <<'PY'
import os

import boto3
from botocore.client import Config
from botocore.exceptions import ClientError

client = boto3.client(
    "s3",
    endpoint_url=os.environ["MINIO_ENDPOINT"],
    aws_access_key_id=os.environ["MINIO_ACCESS_KEY"],
    aws_secret_access_key=os.environ["MINIO_SECRET_KEY"],
    region_name="us-east-1",
    config=Config(signature_version="s3v4", s3={"addressing_style": "path"}),
)

bucket = os.environ["MINIO_BUCKET_NAME"]
key = os.environ["MINIO_OBJECT_KEY"]
body = os.environ["MINIO_FILE_CONTENT"].encode()

try:
    client.head_bucket(Bucket=bucket)
except ClientError:
    client.create_bucket(Bucket=bucket)

client.put_object(Bucket=bucket, Key=key, Body=body)
value = client.get_object(Bucket=bucket, Key=key)["Body"].read().decode()
if value != os.environ["MINIO_FILE_CONTENT"]:
    raise SystemExit(f"unexpected object content: {value!r}")
PY
    rc=$?

    kill ${pf_pid} >/dev/null 2>&1 || true
    wait ${pf_pid} 2>/dev/null || true

    if [[ ${rc} -ne 0 ]]; then
        panic "failed to seed data into ${app_name}"
    fi
}

function setup_minio() {
    kubectl create -f test/gha-e2e/jindo/minio.yaml
    kubectl rollout status --timeout=180s deployment/minio || panic "minio deployment is not ready"
    kubectl rollout status --timeout=180s deployment/minio-a || panic "minio-a deployment is not ready"
    kubectl rollout status --timeout=180s deployment/minio-b || panic "minio-b deployment is not ready"

    seed_minio_bucket minio minioadmin minioadmin mybucket testfile helloworld 19000
    if should_use_minio_multi_oss; then
        seed_minio_bucket minio-a "$multi_oss_access_key_a" "$multi_oss_secret_key_a" "$multi_oss_bucket_a" "$multi_oss_object_key_a" "$multi_oss_expected_data_a" 19001
        seed_minio_bucket minio-b "$multi_oss_access_key_b" "$multi_oss_secret_key_b" "$multi_oss_bucket_b" "$multi_oss_object_key_b" "$multi_oss_expected_data_b" 19002
    fi
}

function setup_oss_emulator() {
    kubectl create -f test/gha-e2e/jindo/oss-emulator.yaml
    kubectl rollout status --timeout=180s deployment/oss-a || panic "oss-a deployment is not ready"
    kubectl rollout status --timeout=180s deployment/oss-b || panic "oss-b deployment is not ready"
}

function create_dataset() {
    dataset_file=$1
    dataset_name=$2

    kubectl create -f $dataset_file

    if [[ -z "$(kubectl get dataset $dataset_name -oname)" ]]; then
        panic "failed to create dataset $dataset_name"
    fi

    if [[ -z "$(kubectl get jindoruntime $dataset_name -oname)" ]]; then
        panic "failed to create jindoruntime $dataset_name"
    fi
}

function wait_dataset_bound() {
    dataset_name=$1
    deadline=180 # 3 minutes
    last_state=""
    log_interval=0
    log_times=0
    while true; do
        last_state=$(kubectl get dataset $dataset_name -ojsonpath='{@.status.phase}')
        if [[ $log_interval -eq 3 ]]; then
            log_times=$(expr $log_times + 1)
            syslog "checking dataset.status.phase==Bound (already $(expr $log_times \* $log_interval \* 5)s, last state: $last_state)"
            if [[ "$(expr $log_times \* $log_interval \* 5)" -ge "$deadline" ]]; then
                panic "timeout for ${deadline}s!"
            fi
            log_interval=0
        fi

        if [[ "$last_state" == "Bound" ]]; then
            break
        fi
        log_interval=$(expr $log_interval + 1)
        sleep 5
    done
    syslog "Found dataset $dataset_name status.phase==Bound"
}

function wait_runtime_stable() {
    runtime_name=$1
    deadline=300
    elapsed=0

    while true; do
        master_phase=$(kubectl get jindoruntime $runtime_name -ojsonpath='{@.status.masterPhase}' 2>/dev/null)
        worker_phase=$(kubectl get jindoruntime $runtime_name -ojsonpath='{@.status.workerPhase}' 2>/dev/null)
        fuse_phase=$(kubectl get jindoruntime $runtime_name -ojsonpath='{@.status.fusePhase}' 2>/dev/null)
        fuse_pod=$(kubectl get pod -l release=$runtime_name,role=jindofs-fuse -ojsonpath='{.items[0].metadata.name}' 2>/dev/null)
        fuse_restart_count=""
        if [[ -n "$fuse_pod" ]]; then
            fuse_restart_count=$(kubectl get pod $fuse_pod -ojsonpath='{@.status.containerStatuses[0].restartCount}' 2>/dev/null)
        fi

        if [[ "$master_phase" == "Ready" && "$worker_phase" == "Ready" && "$fuse_phase" == "Ready" && -n "$fuse_pod" && -n "$fuse_restart_count" ]]; then
            sleep 20
            fuse_restart_count_after=$(kubectl get pod $fuse_pod -ojsonpath='{@.status.containerStatuses[0].restartCount}' 2>/dev/null)
            if [[ "$fuse_restart_count_after" == "$fuse_restart_count" ]]; then
                syslog "Found runtime $runtime_name stable with fuse pod $fuse_pod (restartCount=$fuse_restart_count_after)"
                break
            fi
        fi

        elapsed=$(expr $elapsed + 5)
        if [[ "$elapsed" -ge "$deadline" ]]; then
            panic "timeout waiting for jindoruntime $runtime_name to become stable"
        fi

        sleep 5
    done
}

function wait_runtime_components_ready() {
    runtime_name=$1
    deadline=240
    elapsed=0

    while true; do
        master_phase=$(kubectl get jindoruntime $runtime_name -ojsonpath='{@.status.masterPhase}' 2>/dev/null)
        worker_phase=$(kubectl get jindoruntime $runtime_name -ojsonpath='{@.status.workerPhase}' 2>/dev/null)

        if [[ "$master_phase" == "Ready" && "$worker_phase" == "Ready" ]]; then
            syslog "Found runtime $runtime_name master/worker ready"
            break
        fi

        elapsed=$(expr $elapsed + 5)
        if [[ "$elapsed" -ge "$deadline" ]]; then
            panic "timeout waiting for jindoruntime $runtime_name master/worker to become ready"
        fi

        sleep 5
    done
}

function create_warmup_pod() {
    dataset_name=$1
    warmup_name="${dataset_name}-warmup"

    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: ${warmup_name}
spec:
  restartPolicy: Never
  automountServiceAccountToken: false
  containers:
    - name: warmup
      image: registry-cn-hongkong.ack.aliyuncs.com/acs/smartdata:6.9.1-202509151826
      imagePullPolicy: IfNotPresent
      command: ["/bin/sh", "-c"]
      args:
        - sleep 360
      volumeMounts:
        - mountPath: /data
          name: fluid-vol
  volumes:
    - name: fluid-vol
      persistentVolumeClaim:
        claimName: ${dataset_name}
EOF
}

function delete_warmup_pod() {
    dataset_name=$1
    kubectl delete pod "${dataset_name}-warmup" --ignore-not-found >/dev/null 2>&1 || true
    kubectl wait --for=delete --timeout=180s pod/"${dataset_name}-warmup" >/dev/null 2>&1 || true
}

function wait_warmup_ready() {
    dataset_name=$1
    kubectl wait --for=condition=Ready --timeout=180s pod/${dataset_name}-warmup >/dev/null || panic "warmup pod ${dataset_name}-warmup is not ready"
}

function create_job() {
    job_file=$1
    job_name=$2

    kubectl create -f $job_file

    if [[ -z "$(kubectl get job $job_name -oname)" ]]; then
        panic "failed to create job"
    fi
}

function wait_job_completed() {
    job_name=$1
    while true; do
        succeed=$(kubectl get job $job_name -ojsonpath='{@.status.succeeded}')
        failed=$(kubectl get job $job_name -ojsonpath='{@.status.failed}')
        if [[ "$failed" -ne "0" ]]; then
            kubectl logs job/$job_name --all-containers --tail=-1 >/dev/stderr 2>&1 || true
            panic "job failed when accessing data"
        fi
        if [[ "$succeed" -eq "1" ]]; then
            break
        fi
        sleep 5
    done
    syslog "Found succeeded job $job_name"
}

function cleanup_scenario() {
    dataset_file=$1
    dataset_name=$2
    job_file=$3

    delete_warmup_pod $dataset_name
    kubectl delete -f $job_file --ignore-not-found
    kubectl delete -f $dataset_file --ignore-not-found
    kubectl wait --for=delete --timeout=180s jindoruntime/$dataset_name >/dev/null 2>&1 || true
}

function run_scenario() {
    scenario_name=$1
    dataset_file=$2
    dataset_name=$3
    job_file=$4
    job_name=$5

    syslog "Running scenario: $scenario_name"
    create_dataset $dataset_file $dataset_name
    wait_dataset_bound $dataset_name
    create_warmup_pod $dataset_name
    wait_runtime_stable $dataset_name
    wait_warmup_ready $dataset_name
    create_job $job_file $job_name
    wait_job_completed $job_name
    cleanup_scenario $dataset_file $dataset_name $job_file
}

function dump_env_and_clean_up() {
    for dataset_name in $s3_dataset_name $multi_oss_dataset_name; do
        if kubectl get dataset $dataset_name >/dev/null 2>&1; then
            bash tools/diagnose-fluid-jindo.sh collect --name $dataset_name --namespace default --collect-path ./e2e-tmp/testcase-$dataset_name.tgz
        fi
    done
    syslog "Cleaning up resources for testcase $testname"
    kubectl delete -f test/gha-e2e/jindo/ --ignore-not-found
    if [[ -n "$rendered_dir" && -d "$rendered_dir" ]]; then
        rm -rf "$rendered_dir"
    fi
}

function main() {
    syslog "[TESTCASE $testname STARTS AT $(date)]"
    if is_real_oss_mode; then
        multi_oss_backend="real"
    elif ! should_use_minio_multi_oss && ! should_use_emulator_multi_oss; then
        panic "unsupported JINDO_E2E_MULTI_OSS_BACKEND=${multi_oss_backend}, expected minio or emulator"
    fi

    if ! should_skip_s3_scenario || should_use_minio_multi_oss; then
        setup_minio
    fi
    if is_real_oss_mode; then
        syslog "Using real OSS multi-mount verification as manual reinforcement"
        setup_real_oss
    elif should_use_minio_multi_oss; then
        syslog "Using MinIO-backed multi-mount verification for optional compatibility checks"
        multi_oss_endpoint_a="${JINDO_E2E_OSS_ENDPOINT_A:-minio-a.default.svc.cluster.local:9000}"
        multi_oss_endpoint_b="${JINDO_E2E_OSS_ENDPOINT_B:-minio-b.default.svc.cluster.local:9000}"
    else
        syslog "Using emulator-backed multi-mount verification as the default CI gate"
        setup_oss_emulator
    fi
    render_multi_oss_files
    trap dump_env_and_clean_up EXIT
    if ! should_skip_s3_scenario; then
        run_scenario "single-mount s3 secret" $s3_dataset_file $s3_dataset_name $s3_job_file $s3_job_name
    fi
    run_scenario "multi-mount oss secret projections" $multi_oss_dataset_file $multi_oss_dataset_name $multi_oss_job_file $multi_oss_job_name
    syslog "[TESTCASE $testname SUCCEEDED AT $(date)]"
}

main
