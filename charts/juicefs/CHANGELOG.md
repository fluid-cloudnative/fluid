0.1.0

- Support Kubernetes Orchestration via Fluid

0.2.0

- Change worker from Daemonset to Statefulset

0.2.1

- Add owner Reference

0.2.2

- Add runtime identity information

0.2.3

- Support configurable runtime pod metadata

0.2.4

- Support configurable tieredstore's volume type

0.2.5

- Support configurable resource and privileged

0.2.6

- Make fuse tolerate any taint

0.2.7
- Change podManagementPolicy from OrderedReady to Parallel

0.2.8
- Add volumes and volumeMounts to worker and fuse
 
0.2.9
- Add updateStrategy for fuse

0.2.10
- Set root user in worker & fuse pod

0.2.11
- Support credential key in secret

0.2.12
- Set cache dir in volumes & volumeMounts for worker & fuse

0.2.13
- Add `sidecar.istio.io/inject` to components annotation

0.2.14
- Support subpath quota

0.2.15
- Support mirror buckets