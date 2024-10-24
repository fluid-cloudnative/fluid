#!/bin/bash
set -e

for crdfile in $(find /fluid/crds/*.yaml);
do
  crdshort=${crdfile#*_}
  if [[ $(kubectl get --ignore-not-found -f $crdfile | wc -l) -gt 0 ]]; then
    echo "$crdshort founded, replacing its crd..."
    kubectl replace -f $crdfile
  else
    echo "$crdshort not founded, applying its crd..."
    kubectl create -f $crdfile
  fi
done
