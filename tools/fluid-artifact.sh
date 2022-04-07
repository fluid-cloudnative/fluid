#!/usr/bin/env bash
set -e

function print_usage() 
{
    echo "Example: tools/fluid-artifact.sh ./README.md charts/fluid/fluid"
    echo ""
    echo "Usage:"
    echo "  ./fluid-artifact.sh README_PATH CHART1_PATH [CHART_PATH...]"
    echo "OPTIONS:"
    echo "  -h      print this message"
}

function remove_section()
{
    local section="$1"
    local readme="$2"

    local start=$(grep --max-count=1 -n "$section" "$2" | cut -d ":" -f 1)
    if [[ -z $start ]]; then
        echo "Found no section named \"$section\" in $readme"
        return
    fi
    local end=$(sed -n "$(expr $start + 1),\$p" $readme | grep --max-count=1 -n '^#' | cut -d ":" -f 1)
    if [[ -z $end ]]; then
        sed -i "${start},\$d" $readme
    else
        end=$(expr $end + $start - 1)
        sed -i "${start},${end}d" $readme
    fi
    
    echo "Removed section $1"
}

function helm_package()
{
    local chart_path=$1

    local charts_root=$(dirname $chart_path)

    local tar_filepath=$(helm package -d $charts_root $chart_path | awk '{print $NF}' | awk -F "/" '{print $NF}')

    echo "$chart_path has been helm packaged to $tar_filepath"

    helm repo index $charts_root --merge ${charts_root}/index.yaml

    echo "${charts_root}/index.yaml updated"
}

function archive_charts_and_index()
{
    local charts_root=$(dirname $1)
    cd $charts_root

    tar cvf fluid-artifact.tar 'index.yaml' fluid-*.tgz

    echo "Successfully packaged helm chart and updated index.yaml"
    echo "The archived tar file can be found at ${charts_root}/fluid-artifact.tar"
}

function main() {
    if [[ $# -lt 2 ]]; then 
        print_usage
        exit 1
    fi

    if [[ $1 == "-h" ]]; then
        print_usage
        exit 0
    fi

    README_FILE=$(realpath $1)
    shift

    if [[ ! -f $README_FILE ]]; then
        echo "$README_FILE does not exist"
        exit 1
    fi

    while [[ $# -gt 0 ]]; do
        FLUID_CHART_PATH=$(realpath $1)

        if [[ ! -d $FLUID_CHART_PATH ]]; then
            echo "$FLUID_CHART_PATH does not exist"
            exit 1
        fi
        
        CHART_README=${FLUID_CHART_PATH}/README.md
        echo "Converting ${README_FILE} and store it to ${CHART_README}"
        cat ${README_FILE} > ${CHART_README}

        secs_to_remove=("## Quick Demo" "## Community" "## Adopters")

        IFS=""
        for sec in ${secs_to_remove[*]} 
        do
            remove_section "${sec}" "$CHART_README"
        done

        helm_package "${FLUID_CHART_PATH}"

        # rm ${CHART_README}

        shift
    done

    archive_charts_and_index "$FLUID_CHART_PATH"
}

main "$@"
