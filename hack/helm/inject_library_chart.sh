#!/bin/bash

# USAGE:
#   FLUID_HOME=/path/to/fluid
#   cd $FLUID_HOME
#   bash hack/helm/inject_library_chart.sh

set -e

# All charts should be injected except for Fluid control-plane chart and library chart itself.
charts=$(find charts | grep Chart.yaml | grep -v "charts/fluid/\|charts/library" | xargs -I _ dirname _)

for chart in $charts; do
    relpath=$(realpath charts/library --relative-to $chart/charts)
    echo $relpath
    if [[ ! -d "$chart/charts/library" ]]; then
        mkdir -p "$chart/charts" && ln -s $relpath $chart/charts/library
        echo "Injecting $chart with library chart."
    fi

    # library chart is supported in Helm chart apiVersion=v2.
    if [[ -f "$chart/Chart.yaml" && $(grep -c "dependencies:" $chart/Chart.yaml) == 0 ]]; then
        echo "Processing $chart/Chart.yaml."
        sed -i "" 's/apiVersion: v1/apiVersion: v2/' $chart/Chart.yaml
        cat << EOF >> $chart/Chart.yaml

dependencies:
- name: library
  version: "0.1.0"
  repository: "file://$(realpath charts/library --relative-to $chart)"
EOF
    fi
done