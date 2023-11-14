# Demo - Use pprof to analyze the performance of Fluid components

## Background introduction
pprof is a tool for visualizing and analyzing performance data. It can collect CPU, memory, stack and other information of programs, and generate text and graphical reports.

The Fluid community has enabled the pprof service in each component. Users can access it * http://127.0.0.1:6060/debug/pprof/ * in the component Pod.

## Prerequisites

You can download the latest Fluid installation package from [Fluid Releases](https://github.com/fluid-cloudnative/fluid/releases).

Refer to the [Installation document](../userguide/install.md) to complete the installation.

```shell
$ kubectl get po -n fluid-system
NAME                                         READY   STATUS              RESTARTS   AGE
csi-nodeplugin-fluid-ctc4l                   2/2     Running             0          113s
csi-nodeplugin-fluid-k7cqt                   2/2     Running             0          113s
csi-nodeplugin-fluid-x9dfd                   2/2     Running             0          113s
dataset-controller-57ddd56b54-9vd86          1/1     Running             0          113s
fluid-webhook-84467465f8-t65mr               1/1     Running             0          113s
```

Make sure `dataset-controller`, `fluid-webhook` pod and `csi-nodeplugin` pods work well. `juicefs-runtime-controller` will be installed automatically when JuiceFSRuntime created.

## Enter the component Pod for performance analysis
View the name of each Fluid component Pod (this article uses `Fluid dataset controller` as an example for performance analysis).
```shell
$ kubectl get pods -n fluid-system
NAME                                  READY   STATUS    RESTARTS         AGE
csi-nodeplugin-fluid-kg9bc            2/2     Running   0                22h
csi-nodeplugin-fluid-nbbjk            2/2     Running   0                22h
csi-nodeplugin-fluid-vjdfz            2/2     Running   0                22h
dataset-controller-77cfc8f9bf-m488j   1/1     Running   0                22h
fluid-webhook-5f76bb6567-jwpbk        1/1     Running   0                22h
fluidapp-controller-b7c4d5579-ztvlw   1/1     Running   0                22h
```
Enter the dataset-controller pod.
```shell
$ kubectl exec -it dataset-controller-77cfc8f9bf-m488j -n fluid-system bash
```


## Perform performance analysis
After installing the Go environment, the **go tool pprof** command can be used for perform performance analysis. In addition, users can also access *http://127.0.0.1:6060/debug/pprof/* to view some data information.

The following data of the program can be analyzed：
- allocs：A sampling of all past memory allocations
- block：Stack traces that led to blocking on synchronization primitives
- cmdline： The command line invocation of the current program。
- goroutine：Stack traces of all current goroutines
- heap：A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.
- mutex：Stack traces of holders of contended mutexes
- profile： CPU profile. You can specify the duration in the seconds GET parameter. After you get the profile file, use the go tool pprof command to investigate the profile.
- threadcreate：Stack traces that led to the creation of new OS threads
- trace：A trace of execution of the current program. You can specify the duration in the seconds GET parameter. After you get the trace file, use the go tool trace command to investigate the trace.


Collect the data you are interested in. This article takes the 30 second CPU data as an example and saves the data as a `profile.out`. You can use the `go tool pporf` command for analysis with the `profile.out` locally or on the host (the Go environment is required).

```shell
$ curl -o profile.out http://localhost:6060/debug/pprof/profile?seconds=30
$ go tool pprof test.out 
File: dataset-controller
Type: cpu
Time: Nov 2, 2022 at 6:48pm (CST)
Duration: 29.91s, Total samples = 50ms ( 0.17%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) 
```
Enter the `top5` command in the interactive terminal to print the top 5 functions that consume CPU resources.
```shell
(pprof) top5   
Showing nodes accounting for 50ms, 100% of 50ms total
Showing top 5 nodes out of 64
      flat  flat%   sum%        cum   cum%
      10ms 20.00% 20.00%       10ms 20.00%  github.com/fluid-cloudnative/fluid/vendor/golang.org/x/net/http2.(*Transport).expectContinueTimeout
      10ms 20.00% 40.00%       10ms 20.00%  net/http.cloneURL
      10ms 20.00% 60.00%       10ms 20.00%  path.(*lazybuf).append
      10ms 20.00% 80.00%       10ms 20.00%  runtime.memclrNoHeapPointers
      10ms 20.00%   100%       10ms 20.00%  runtime.newarray
```

**For more usage information, please refer to [Document](https://github.com/google/pprof)**