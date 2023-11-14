#!/bin/bash

set +x

print_usage()
{
  echo -e "Usage:"
  echo -e "  ./job-summary.sh [options]."
  echo -e "OPTIONS:"
  echo -e "  -h \t Display this help message."
  echo -e "  --job \t Set the name of job."
  echo -e "  --log \t Set the path of logs file."
  echo -e "  --laucher \t Set the path of job laucher desctiption file."
  echo -e "EXAMPLE:"
  echo -e "  ./job-summary.sh --job <JOB_NAME>"
  echo -e "  ./job-summary.sh --log <PATH_TO_JOB_LOG> --laucher <PATH_TO_LAUCHER_DESCRIPTION>"
}

training_GPUs()
{
  gpus=$(egrep "^1\simages/sec" "${log_path}" | wc -l)
  echo "${gpus}"
}

training_steps()
{
  total_steps=$(tail -1000 "${log_path}" | grep "images/sec" | tail -1 | awk '{print $1}')
  echo "${total_steps}"
}

training_accuracy()
{
  accuracy=$(tail -1000 "${log_path}" | grep "Accuracy @ 5" | tail -1 | awk '{print $10}')
  echo "${accuracy}"
}

training_speed_at_step()
{
  local step=${1}
  speed=$(egrep "^${step}\simages/sec" "${log_path}" | \
    awk '{cnt+=1; total_speed+=$3}; END {avg_speed=(total_speed/(cnt+0.00001)); print total_speed,avg_speed}')
  echo "${speed}"
}

training_average_speed_until_steps()
{
  steps="${1}"
  avg_speed=$(grep "images/sec" "${log_path}" | \
    awk '$1<=steps {n+=1; total+=$3}; END{avg_gpu=total/(n+0.000001);avg_step=total*gpus/(n+0.000001); print avg_step,avg_gpu,n}' gpus=${gpus} steps=${steps})
  echo "${avg_speed}"
}

training_average_speed_from_to_steps()
{
  lower="${1}"
  upper="${2}"
  avg_speed=$(grep "images/sec" "${log_path}" | \
    awk '$1<=upper&&$1>=lower {n+=1; total+=$3}; END{avg_gpu=total/(n+0.000001);avg_step=total*gpus/(n+0.000001); print avg_step,avg_gpu,n}' gpus=${gpus} steps=${steps} lower=${lower} upper=${upper})
  echo "${avg_speed}"
}

training_average_speed_of_all_steps()
{
  avg_speed=$(grep "images/sec" "${log_path}" | \
    awk '{n+=1; total+=$3}; END{avg_gpu=total/(n+0.000001);avg_step=total*gpus/(n+0.000001); print avg_step,avg_gpu,n}' gpus=${gpus})
  echo "${avg_speed}"
}

training_average_speed()
{
  local sum=0
  local tmp_log="/tmp/${job_name}_$(date '+%Y%m%d%H%M')"
  arena logs ${job_name} > ${tmp_log}
  # average speed
  for i in $(seq 0 10 2000); do
    sum=$(egrep "^${i}\simages/sec" ${tmp_log} | awk '{cnt+=1; step_sum+=$3}; END {sum+=(step_sum/(cnt+0.00001)) ;print int(sum)}' sum=${sum})
    if [ ${i} -eq 100 ] || [ ${i} -eq 500 ] || [ ${i} -eq 1000 ] || [ ${i} -eq 2000 ]; then
      avg=$((sum * 10 / i))
      echo -e "Top ${i}:\t${avg} images/sec"
    fi
  done
}

get_pod_timestamp()
{
  local pod_path=$1
  local status=$2
  local tf=$(egrep "^Containers" -A20 ${pod_path} \
    | grep ${status} \
    | awk '{print $2,$3,$4,$5,$6,$7}' \
    | xargs -I {} date "+%s" -d {})
  echo ${tf}
}

compute_pod_lifetime()
{
  local pod_path=$1
  local started=$(get_pod_timestamp ${pod_path} "Started")
  local finished=$(get_pod_timestamp ${pod_path} "Finished")
  echo $((${finished} - ${started}))
}

training_time()
{
  # get laucher
  local seconds=$(compute_pod_lifetime ${laucher_path})
  local duration="0s"
  if [ ${seconds} -lt 60 ]; then
    duration="${seconds}s"
  elif [ ${seconds} -lt 3600 ]; then
    duration="$((${seconds}/60))m$((${seconds}%60))s"
  else
    duration="$((${seconds}/3600))h$(((${seconds}%3600)/60))m$((${seconds}%60))s"
  fi
  echo "${duration}"
}

summary()
{
  gpus=$(training_GPUs)
  total_steps=$(training_steps)
  startup_step=1000

  # average speed of 25%
  step_25=$((total_steps / 10 / 4 * 10))
  avg_speed_25=$(training_average_speed_until_steps ${step_25})

  # average speed of 50%
  step_50=$((total_steps / 10 / 4 * 10 * 2))
  avg_speed_50=$(training_average_speed_until_steps ${step_50})

  # average speed of 75%
  step_75=$((total_steps / 10 / 4 * 10 * 3))
  avg_speed_75=$(training_average_speed_until_steps ${step_75})

  # 100%
  avg_speed_100=$(training_average_speed_of_all_steps)

  ########## test
  avg_speed_25_to_50=$(training_average_speed_from_to_steps ${step_25} ${step_50})
  avg_speed_50_to_75=$(training_average_speed_from_to_steps ${step_50} ${step_75})
  avg_speed_75_to_100=$(training_average_speed_from_to_steps ${step_75} ${total_steps})

  echo "==============SUMMARY=================="
  echo -e "Name: \t\t ${job_name}"
  echo -e "Duration: \t\t $(training_time)"
  echo -e "GPUs: \t\t ${gpus}"
  echo -e "Steps: \t\t $(training_steps)"
  echo -e "Accuracy@5: \t\t $(training_accuracy)"
  echo -e "Speed of \t\t\t Step \t GPU \t cnt"
  echo -e "Speed@${startup_step}: \t\t\t $(training_speed_at_step ${startup_step})"
  echo -e "Speed@${total_steps}: \t\t\t $(training_speed_at_step ${total_steps})"
  echo -e "Average Speed 25%(${step_25}): \t ${avg_speed_25}"
  echo -e "Average Speed 50%(${step_50}): \t ${avg_speed_50}"
  echo -e "Average Speed 75%(${step_75}): \t ${avg_speed_75}"
  echo -e "Average Speed 100%(${total_steps}): \t ${avg_speed_100}"
  echo -e "Average Speed 0% to 25%: \t ${avg_speed_25}"
  echo -e "Average Speed 25% to 50%: \t ${avg_speed_25_to_50}"
  echo -e "Average Speed 50% to 75%: \t ${avg_speed_50_to_75}"
  echo -e "Average Speed 75% to 100%: \t ${avg_speed_75_to_100}"
  echo "================END===================="
}

ensure_params()
{
  if [[ -z "${log_path}" ]]; then
    log_path="/tmp/${job_name}_$(date '+%Y%m%d%H%M').log"
    arena logs ${job_name} &>"${log_path}" 2>&1
  fi

  if [[ -z "${laucher_path}" ]]; then
    local job_laucher=$(arena get ${job_name} \
      | grep "launcher" | awk '{print $5}')
    laucher_path="/tmp/${job_laucher}_$(date '+%Y%m%d%H%M').log"
    kubectl describe po ${job_laucher} &>"${laucher_path}" 2>&1
  fi
}

main()
{
  # Parse arguments using getopt
  ARGS=$(getopt -a -o h --long help,job:,,log:,laucher: -- "$@")
  if [ $? != 0 ]; then
    exit 1
  fi

  eval set -- "${ARGS}"

  while true
  do
    case "$1" in
      -h|--help)
        print_usage
        shift 1
        exit 0
        ;;
      --job)
        job_name=$2
        shift 2
        ;;
      --log)
        log_path=$2
        shift 2
        ;;
      --laucher)
        laucher_path=$2
        shift 2
        ;;
      --)
        shift
        break
        ;;
      *)
        echo "ERROR: invalide argument $1" >&2
        exit 1
        ;;
    esac
  done

  if [[ -z "${job_name}" && -z "${log_path}" && -z "${laucher_path}" ]]; then
    echo "ERROR: invalide aruguments" >&2
    print_usage
    exit 1
  fi

  ensure_params
  summary
}

main "$@"