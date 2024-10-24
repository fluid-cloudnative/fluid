kubectl scale deployment alluxioruntime-controller --replicas=0 -nfluid-system
kubectl delete -f dataset.yaml
kubectl delete -f runtime.yaml