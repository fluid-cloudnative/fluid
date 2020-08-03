# fluid诊断脚本使用说明

## 脚本介绍

fluid提供了shell脚本[diagnose-fluid.sh](../../scripts/diagnose-fluid.sh)帮助用户快速收集fluid系统和Runtime容器的日志信息。

## 如何使用

首先，确保shell脚本有运行权限

```bash
$ chmod a+x diagnose-fluid.sh
```

运行`diagnose-fluid.sh`，`--name`指定了Runtime的name，`--namespace`指定了Runtime的namespace

```bash
$ ./diagnose-fluid.sh --name cifar10 --namespace default
```

shell脚本会将收集的日志信息打包到执行路径下的一个压缩包里。

