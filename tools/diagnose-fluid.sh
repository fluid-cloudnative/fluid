#!/usr/bin/env bash
set +x

print_usage() {
  echo "Usage:"
  echo "    ./diagnose-fluid.sh COMMAND [OPTIONS]"
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
  echo "    -t, --type name"
  echo "        Set the type of runtime. Current avaliable type: alluxio, goosefs, jindo, juicefs."
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

pod_status() {
  local namespace=${1:-"default"}
  run kubectl get po -owide -n ${namespace} &>"$diagnose_dir/pods-${namespace}.log"
}

fluid_pod_logs() {
  core_component "${fluid_namespace}" "manager" "control-plane=${runtime_type}runtime-controller"
  core_component "${fluid_namespace}" "manager" "control-plane=dataset-controller"
  core_component "${fluid_namespace}" "plugins" "app=csi-nodeplugin-fluid"
  core_component "${fluid_namespace}" "node-driver-registrar" "app=csi-nodeplugin-fluid"
}

alluxioruntime_pod_logs() {
  core_component "${runtime_namespace}" "alluxio-master" "role=alluxio-master" "release=${runtime_name}"
  core_component "${runtime_namespace}" "alluxio-job-master" "role=alluxio-master" "release=${runtime_name}"
  core_component "${runtime_namespace}" "alluxio-worker" "role=alluxio-worker" "release=${runtime_name}"
  core_component "${runtime_namespace}" "alluxio-job-worker" "role=alluxio-worker" "release=${runtime_name}"
  core_component "${runtime_namespace}" "alluxio-fuse" "role=alluxio-fuse" "release=${runtime_name}"
}

goosefsruntime_pod_logs() {
  core_component "${runtime_namespace}" "goosefs-master" "role=goosefs-master" "release=${runtime_name}"
  core_component "${runtime_namespace}" "goosefs-job-master" "role=goosefs-master" "release=${runtime_name}"
  core_component "${runtime_namespace}" "goosefs-worker" "role=goosefs-worker" "release=${runtime_name}"
  core_component "${runtime_namespace}" "goosefs-job-worker" "role=goosefs-worker" "release=${runtime_name}"
  core_component "${runtime_namespace}" "goosefs-fuse" "role=goosefs-fuse" "release=${runtime_name}"
}

jindoruntime_pod_logs() {
  core_component "${runtime_namespace}" "jindofs-master" "role=jindofs-master" "release=${runtime_name}"
  core_component "${runtime_namespace}" "jindofs-worker" "role=jindofs-worker" "release=${runtime_name}"
  core_component "${runtime_namespace}" "jindofs-fuse" "role=jindofs-fuse" "release=${runtime_name}"
}

juicefsruntime_pod_logs() {
  core_component "${runtime_namespace}" "juicefs-worker" "role=juicefs-worker" "release=${runtime_name}"
  core_component "${runtime_namespace}" "juicefs-fuse" "role=juicefs-fuse" "release=${runtime_name}"
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
  kubectl describe ${runtime_type}runtime --namespace ${runtime_namespace} ${name} &>"${diagnose_dir}/${runtime_type}runtime-${runtime_name}.yaml" 2>&1
  kubectl describe pv ${runtime_namespace}-${runtime_name} &>"${diagnose_dir}/pv-${runtime_name}.yaml" 2>&1
  kubectl describe pvc ${runtime_name} --namespace ${runtime_namespace} &>"${diagnose_dir}/pvc-${runtime_name}.yaml" 2>&1
}

archive() {
  tar -zcvf "${current_dir}/diagnose_fluid_${timestamp}.tar.gz" "${diagnose_dir}"
  echo "please get diagnose_fluid_${timestamp}.tar.gz for diagnostics"
}

pd_collect() {
  echo "Start collecting, runtime-type=${runtime_type}, runtime-name=${runtime_name}, runtime-namespace=${runtime_namespace}"
  helm_get "${fluid_name}"
  helm_get "${runtime_name}"
  pod_status "${fluid_namespace}"
  pod_status "${runtime_namespace}"
  ${runtime_type}runtime_pod_logs
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
  runtime_type=${runtime_type:?"the type of runtime must be set"}
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
      -t|--type)
        runtime_type=$2
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

  if [[ "${runtime_type}" != "alluxio" ]] && [[ "${runtime_type}" != "goosefs" ]] && [[ "${runtime_type}" != "jindo" ]] && [[ "${runtime_type}" != "juicefs" ]]; then
    echo "Wrong runtime type."
    print_usage
    return
  fi

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
