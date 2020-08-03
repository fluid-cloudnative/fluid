#!/usr/bin/env bash
set +x

# arguments
fluid_name="fluid"
fluid_namespace="fluid-system"
runtime_name="imagenet"
runtime_namespace="default"

current_dir=$(pwd)
timestamp=$(date +%s)
diagnose_dir=/tmp/diagnose_fluid_${timestamp}
mkdir -p $diagnose_dir

print_usage() {
  echo "Usage:"
  echo "    ./diagnose-fluid.sh [options]"
  echo "OPTIONS:"
  echo "    -h, --help"
  echo "        Display this help message."
  echo "    --name name"
  echo "        Set the name of runtime (default '${runtime_name}')."
  echo "    --namespace name"
  echo "        Set the namespace of runtime (default '${runtime_namespace}')."
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

helm_status() {
  run helm status ${fluid_name} &>"$diagnose_dir/helm.log"
}

pod_status() {
  local namespace=${1:=default}
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
  local selectors="$@"
  local constrains=$(echo "${selectors}" | tr ' ' ',')
  if [[ -n ${constrains} ]]; then
    constrains="-l ${constrains}"
  fi
  mkdir -p "$diagnose_dir/pods-${namespace}"
  local pods=$(kubectl get po -n ${namespace} ${constrains} | awk '{print $1}' | grep -v NAME)
  for po in ${pods}; do
    kubectl logs ${po} -c $container -n ${namespace} 2>&1 &>"$diagnose_dir/pods-${namespace}/${po}-${container}.log"
  done
}

archive() {
  tar -zcvf ${current_dir}/diagnose_fluid_${timestamp}.tar.gz ${diagnose_dir}
  echo "please get diagnose_fluid_${timestamp}.tar.gz for diagnostics"
}

pd_collect() {
  helm_status
  pod_status ${fluid_namespace}
  pod_status ${runtime_namespace}
  fluid_pod_logs
  runtime_pod_logs
  archive
}

main() {
  # Parse arguments using getopt
  ARGS=$(getopt -a -o h --long help,name:,namespace: -- "$@")
  if [ $? != 0 ]; then
    exit 1
  fi

  eval set -- "${ARGS}"

  while true; do
    case "$1" in
    -h | --help)
      print_usage
      shift 1
      exit 0
      ;;
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

  pd_collect
}

main "$@"
