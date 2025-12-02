#!/bin/bash

# USAGE:
#   FLUID_HOME=/path/to/fluid
#   cd $FLUID_HOME
#   bash hack/helm/pin_runtime_chart_version.sh <fluid_version>

set -e

fluid_version=$1
if [ -z "$fluid_version" ]; then
  echo "$0 got an error: please specify a non-empty Fluid version to continue"
  exit 1
fi

# All charts should be injected except for Fluid control-plane chart and library chart itself.
charts=$(find charts | grep Chart.yaml | grep -v "charts/fluid/\|charts/library" | xargs -I _ dirname _)

for chart in $charts; do
    if [[ -f "$chart/Chart.yaml" ]]; then
        echo "Processing $chart/Chart.yaml."
        if [[ $(grep -c "appVersion:" $chart/Chart.yaml) == 0 ]]; then
            echo "appVersion: $fluid_version" >> $chart/Chart.yaml
        else
            sed -i "s|^appVersion:.*|appVersion: ${fluid_version//\//\\/}|" "$chart/Chart.yaml"
        fi
    fi
done
