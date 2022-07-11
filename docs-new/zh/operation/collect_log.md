# Fluid问题诊断

您可能会在部署、开发Fluid的过程中遇到各种问题，而查看日志可以协助我们定位问题原因。但在分布式环境下，Fluid底层的分布式缓存引擎（Runtime）运行在不同主机的容器上，手动收集这些容器的日志效率低下。因此，Fluid提供了shell脚本[diagnose-fluid.sh](https://raw.githubusercontent.com/fluid-cloudnative/fluid/master/tools/diagnose-fluid.sh)，帮助使用者快速收集Fluid系统和Runtime容器的日志信息。

## 如何使用脚本收集日志

1. 首先，确保shell脚本有运行权限
    ```bash
    $ chmod a+x diagnose-fluid.sh
    ```
   
2. 查看帮助信息

    ```bash
    $ ./diagnose-fluid.sh 
    Usage:
        ./diagnose-fluid.sh COMMAND [OPTIONS]
    COMMAND:
        help
            Display this help message.
        collect
            Collect pods logs of controller and runtime.
    OPTIONS:
        -r, --name name
            Set the name of runtime.
        -n, --namespace name
            Set the namespace of runtime.
    ```

3. 收集日志

    运行`diagnose-fluid.sh`，`--name`指定了Runtime的name，`--namespace`指定了Runtime的namespace
    
    ```bash
    $ ./diagnose-fluid.sh collect --name cifar10 --namespace default
    ```
    
    shell脚本会将收集的日志信息打包到执行路径下的一个压缩包里。
