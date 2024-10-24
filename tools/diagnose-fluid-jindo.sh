#!/usr/bin/env bash
set +x

print_usage() {
  echo "Usage:"
  echo "    ./diagnose-fluid-jindo.sh COMMAND [OPTIONS]"
  echo "COMMAND:"
  echo "    help"
  echo "        Display this help message."
  echo "    collect"
  echo "        Collect pods logs of controller and runtime."
  echo "OPTIONS:"
  echo "    -r, --name name"
  echo "        Set the name of runtime."
  echo "    -n, --namespace name"
  echo "        Set the namespace of runtime."
  echo "    --collect-path"
  echo "        Set which file the information is collected into. (default: $(pwd)/diagnose_fluid_\${timestamp}.tar.gz)"
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
  run helm get all -n ${runtime_namespace} "${1}" &>"$diagnose_dir/helm-${1}.yaml"
}

helm_get_runtime() {
  run env HELM_DRIVER=configmap helm get all -n ${runtime_namespace} "${1}" &>"$diagnose_dir/helm-${1}.yaml"
}

pod_status() {
  local namespace=${1:-"default"}
  run kubectl get po -owide -n ${namespace} &>"$diagnose_dir/pods-${namespace}.log"
  run kubectl get po -oyaml -n ${namespace} &>>"$diagnose_dir/pods-${namespace}.log"
}

fluid_pod_logs() {
  core_component "${fluid_namespace}" "manager" "control-plane=jindoruntime-controller"
  core_component "${fluid_namespace}" "manager" "control-plane=dataset-controller"
  core_component "${fluid_namespace}" "plugins" "app=csi-nodeplugin-fluid"
  core_component "${fluid_namespace}" "node-driver-registrar" "app=csi-nodeplugin-fluid"
}

runtime_pod_logs() {
  core_component "${runtime_namespace}" "jindofs-master" "role=jindofs-master" "release=${runtime_name}"
  core_component "${runtime_namespace}" "jindofs-worker" "role=jindofs-worker" "release=${runtime_name}"
  core_component "${runtime_namespace}" "jindofs-fuse" "role=jindofs-fuse" "release=${runtime_name}"
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

kubectl_resource() {
  # runtime, dataset, pv and pvc should have the same name
  kubectl describe dataset --namespace ${runtime_namespace} ${runtime_name} &>"${diagnose_dir}/dataset-${runtime_name}.yaml" 2>&1
  kubectl describe jindoruntime --namespace ${runtime_namespace} ${name} &>"${diagnose_dir}/jindoruntime-${runtime_name}.yaml" 2>&1
  kubectl describe pv ${runtime_namespace}-${runtime_name} &>"${diagnose_dir}/pv-${runtime_name}.yaml" 2>&1
  kubectl describe pvc ${runtime_name} --namespace ${runtime_namespace} &>"${diagnose_dir}/pvc-${runtime_name}.yaml" 2>&1
}

archive() {
  tar_filename="${current_dir}/diagnose_fluid_${timestamp}.tar.gz"
  if [[ ! -z "${collect_path}" ]]; then
    tar_filename=${collect_path}
    mkdir -p $(dirname "$tar_filename")
  fi
  tar -zcvf "${tar_filename}" "${diagnose_dir}"
  echo "please get ${tar_filename} for diagnostics"
}

pd_collect() {
  echo "Start collecting, runtime-name=${runtime_name}, runtime-namespace=${runtime_namespace}"
  helm_get "${fluid_name}"
  helm_get_runtime "${runtime_name}"
  pod_status "${fluid_namespace}"
  pod_status "${runtime_namespace}"
  runtime_pod_logs
  fluid_pod_logs
  kubectl_resource
  archive
}

collect()
{
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
  if [[ $# -eq 0 ]]; then
    print_usage
    exit 1
  fi

  action="help"

  while [[ $# -gt 0 ]]; do
    case $1 in
      -h|--help|"-?")
        print_usage
        exit 0;
        ;;
      collect|help)
        action=$1
        ;;
      -r|--name)
        runtime_name=$2
        shift
        ;;
      -n|--namespace)
        runtime_namespace=$2
        shift
        ;;
      --collect-path)
        collect_path=$2
        shift
        ;;
      *)
        echo  "Error: unsupported option $1" >&2
        print_usage
        exit 1
        ;;
    esac
    shift
  done

  case ${action} in
    collect)
      collect
      ;;
    help)
      print_usage
      ;;
  esac
}

main "$@"
