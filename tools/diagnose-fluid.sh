#!/usr/bin/env bash
set +x

print_usage() {
  echo "Usage:"
  echo "    ./diagnose-fluid.sh COMMAND [OPTIONS]"
  echo "COMMAND:"
  echo "    help"
  echo "        Display this help message."
  echo "    collect"
  echo "        Collect pods logs of Runtime."
  echo "OPTIONS:"
  echo "    --name name"
  echo "        Set the name of runtime."
  echo "    --namespace name"
  echo "        Set the namespace of runtime."
}

run() {
  echo
  echo "-----------------run $*------------------"
  timeout 10s "$@"
  if [ $? != 0 ]; then
    echo "failed to collect info: $*"
  fi
  echo "------------End of ${1}----------------"
}

helm_get() {
  run helm get all "${1}" &>"$diagnose_dir/${1}.yaml"
}

pod_status() {
  local namespace=${1:-"default"}
  run kubectl get po -owide -n ${namespace} &>"$diagnose_dir/${namespace}.log"
}

fluid_pod_logs() {
  core_component "${fluid_namespace}" "manager" "control-plane=controller-manager"
  core_component "${fluid_namespace}" "plugins" "app=csi-nodeplugin-fluid"
  core_component "${fluid_namespace}" "node-driver-registrar" "app=csi-nodeplugin-fluid"
}

runtime_pod_logs() {
  core_component "${runtime_namespace}" "alluxio-master" "role=alluxio-master" "release=${runtime_name}"
  core_component "${runtime_namespace}" "alluxio-job-master" "role=alluxio-master" "release=${runtime_name}"
  core_component "${runtime_namespace}" "alluxio-worker" "role=alluxio-worker" "release=${runtime_name}"
  core_component "${runtime_namespace}" "alluxio-job-worker" "role=alluxio-worker" "release=${runtime_name}"
  core_component "${runtime_namespace}" "alluxio-fuse" "role=alluxio-fuse" "release=${runtime_name}"
}

core_component() {
  # namespace container selectors...
  local namespace="$1"
  local container="$2"
  shift 2
  local selectors="$*"
  local constrains
  local pods
  constrains=$(echo "${selectors}" | tr ' ' ',')
  if [[ -n ${constrains} ]]; then
    constrains="-l ${constrains}"
  fi
  mkdir -p "$diagnose_dir/pods-${namespace}"
  pods=$(kubectl get po -n ${namespace} "${constrains}" | awk '{print $1}' | grep -v NAME)
  for po in ${pods}; do
    kubectl logs "${po}" -c "$container" -n ${namespace} &>"$diagnose_dir/pods-${namespace}/${po}-${container}.log" 2>&1
  done
}

archive() {
  tar -zcvf "${current_dir}/diagnose_fluid_${timestamp}.tar.gz" "${diagnose_dir}"
  echo "please get diagnose_fluid_${timestamp}.tar.gz for diagnostics"
}

pd_collect() {
  echo "Start collecting, runtime-name=${runtime_name}, runtime-namespace=${runtime_namespace}"
  helm_get "${fluid_name}"
  helm_get "${runtime_name}"
  pod_status "${fluid_namespace}"
  pod_status "${runtime_namespace}"
  runtime_pod_logs
  fluid_pod_logs
  archive
}

collect()
{
  # Parse arguments using getopt
  ARGS=$(getopt -a -o h --long help,name:,namespace: -- "$@")
  if [ $? != 0 ]; then
    exit 1
  fi

  eval set -- "${ARGS}"

  while true; do
    case "$1" in
    --name)
      runtime_name=$2
      shift 2
      ;;
    --namespace)
      runtime_namespace=$2
      shift 2
      ;;
    --)
      shift
      break
      ;;
    *)
      echo "ERROR: invalid argument $1" >&2
      exit 1
      ;;
    esac
  done

  # ensure params
  fluid_name=${fluid_name:-"fluid"}
  fluid_namespace=${fluid_namespace:-"fluid-system"}
  runtime_name=${runtime_name:?"the name of runtime must be set"}
  runtime_namespace=${runtime_namespace:-"default"}

  current_dir=$(pwd)
  timestamp=$(date +%s)
  diagnose_dir="/tmp/diagnose_fluid_${timestamp}"
  mkdir -p "$diagnose_dir"

  pd_collect
}

main() {
  if [[ $# == 0 ]]; then
    print_usage
    exit 1
  fi

  command="$1"
  shift

  case ${command} in
    "collect")
      collect "$@"
      ;;
    "help")
      print_usage
      exit 0
      ;;
    *)
      echo  "ERROR: unsupported command ${command}" >&2
      print_usage
      exit 1
      ;;
  esac
}

main "$@"
