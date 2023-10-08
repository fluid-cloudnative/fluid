0.7.0

Support logConfig <br/>
Support fuse lazy start <br/>
Support fuse critical pod <br/>

0.8.0

Change worker from Daemonset to Statefulset

0.8.1

Repair fuse MountPoint leak issue

0.8.2

Change podManagementPolicy from OrderedReady to Parallel

0.8.3

Add owner Reference

0.8.4

Support more Posix

0.8.5

Add mountPropagation for registrar <br/>
Add auto fuse recovery

0.8.6

Add runtime identity information to template

0.8.7

Fix resources issue 

0.8.8

Fix fuse memleak

0.8.9

Add updateStrategy for fuse

0.8.10

Support configurable tieredstore's volume type

0.8.11

Support configurable pod metadata

0.8.12

Support emptyDir volume source

0.8.13

Repair preRead and meta access issue

0.8.14

Make fuse tolerate any taint

0.8.15

Add Cluster domain 

0.8.16

Use datasetName.datasetNamespace for service discovery

0.8.17

Add pvc storage type support

0.8.18

Add `sidecar.istio.io/inject` to components annotation

0.8.19
Add env variables to components

0.8.20
Add support ak with secret file

0.8.21
Add support dataload with filter

0.8.23
Add support for pvc subpath feature

0.8.24
Fix jindofsx engine sts token volume mount bug