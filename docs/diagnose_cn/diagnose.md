# fluid诊断脚本使用说明

## 脚本介绍

fluid提供了shell脚本[diagnose-fluid.sh](../../tools/diagnose-fluid.sh)帮助用户快速收集fluid系统和Runtime容器的日志信息。

## 如何使用

首先，确保shell脚本有运行权限

```bash
$ chmod a+x diagnose-fluid.sh
```

### 查看使用帮助

```bash
$ ./diagnose-fluid.sh 
Usage:
    ./diagnose-fluid.sh COMMAND [OPTIONS]
COMMAND:
    help
        Display this help message.
    collect
        Collect pods logs of Runtime.
OPTIONS:
    --name name
        Set the name of runtime (default 'imagenet').
    --namespace name
        Set the namespace of runtime (default 'default').
    -a, --all
        Also collect fluid system logs.
```

### 收集日志

运行`diagnose-fluid.sh`，`--name`指定了Runtime的name，`--namespace`指定了Runtime的namespace

```bash
$ ./diagnose-fluid.sh collect --name cifar10 --namespace default
```

shell脚本会将收集的日志信息打包到执行路径下的一个压缩包里。

如果要同时收集`fluid-system`的日志，运行时请额外添加option `--all`

```bash
$ ./diagnose-fluid.sh collect --name cifar10 --namespace default --all
```
