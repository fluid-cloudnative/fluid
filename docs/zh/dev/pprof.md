# 示例 - 使用 pprof 对 fluid 组件进行性能分析

## 背景介绍

pprof 是一个用于可视化和分析性能数据的工具，它可以收集程序的 CPU、内存、堆栈等信息，并对其生成文本和图形报告。

Fluid 社区已经在各个组件内开启了 pprof 服务，用户可以在进入组件 Pod 内部后访问 *http://127.0.0.1:6060/debug/pprof/* 即可看到相关信息。

## 前提条件

在运行该示例之前，请参考[安装文档](../userguide/install.md)完成安装，并检查Fluid各组件正常运行：
```shell
$ kubectl get pod -n fluid-system
alluxioruntime-controller-5b64fdbbb-84pc6   1/1     Running   0          8h
csi-nodeplugin-fluid-fwgjh                  2/2     Running   0          8h
csi-nodeplugin-fluid-ll8bq                  2/2     Running   0          8h
dataset-controller-5b7848dbbb-n44dj         1/1     Running   0          8h
```

通常来说，你会看到一个名为`dataset-controller`的Pod、一个名为`alluxioruntime-controller`的Pod和多个名为`csi-nodeplugin`的Pod正在运行。其中，`csi-nodeplugin`这些Pod的数量取决于你的Kubernetes集群中结点的数量。

## 进入需要进行性能分析的组件 Pod 内部
查看 Fluid 各个组件 Pod 的名称（本文以 Fluid 的 dataset-controller 为例进行性能分析）。
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
进入 dataset-controller Pod 内部。
```shell
$ kubectl exec -it dataset-controller-77cfc8f9bf-m488j -n fluid-system bash
```


## 进行性能分析
在安装完 Go 环境后，可以使用 **go tool pprof** 命令进行性能分析。此外用户也可以访问 *http://127.0.0.1:6060/debug/pprof/* 查看部分数据信息。

pprof 可以分析程序的以下数据：
- allocs：查看过去所有内存分配的样本
- block：查看导致阻塞同步的堆栈跟踪
- cmdline： 当前程序的命令行的完整调用路径
- goroutine：查看当前所有运行的 goroutines 堆栈跟踪
- heap：查看活动对象的内存分配情况
- mutex：查看导致互斥锁的竞争持有者的堆栈跟踪
- profile：CPU 概要文文件
- threadcreate：查看创建新 OS 线程的堆栈跟踪
- trace：当前程序执行的跟踪

- allocs：A sampling of all past memory allocations
- block：Stack traces that led to blocking on synchronization primitives
- cmdline： The command line invocation of the current program。
- goroutine：Stack traces of all current goroutines
- heap：A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.
- mutex：Stack traces of holders of contended mutexes
- profile： CPU profile. You can specify the duration in the seconds GET parameter. After you get the profile file, use the go tool pprof command to investigate the profile.
- threadcreate：Stack traces that led to the creation of new OS threads
- trace：A trace of execution of the current program. You can specify the duration in the seconds GET parameter. After you get the trace file, use the go tool trace command to investigate the trace.

收集您关注的类型的数据，本文以 30 秒 CPU 数据为例，将数据保存为 profile.out。您可以在 Pod 内或将 profile.out 复制到本地使用 `go tool pprof` 命令进行分析（需要 Go 环境）。

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
在交互终端内输入 `top5` 命令打印消耗 CPU 资源前5的函数。
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

**更多使用信息，请查看[文档](https://github.com/google/pprof)**