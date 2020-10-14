# Troubleshooting

You may encounter various problems during installation or development in Fluid. Usually, logs are useful for debugging. But the Runtime containers where Fluid's underlying Distributed Cache Engine is running, are distributed on different hosts under distributed environment, so it's quite annoying to collect these logs one by one. To make this troublesome work easier, we provided a [shell script](https://raw.githubusercontent.com/fluid-cloudnative/fluid/master/tools/diagnose-fluid.sh) to help users collect logs more quickly. This document describes how to use that script.

## Diagnose Fluid using Script

1. Make sure that script is executable
   
   ```shell
   $ chmod a+x diagnose-fluid.sh
   ```

2. Get help message

   ```shell
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

3. Collect logs

   You can collect all the Runtime container logs for given name and namespace with:

   ```shell
   $ ./diagnose-fluid.sh collect --name cifar10 --namespace default
   ```

   > **NOTES**:
   >
   > As you can see from above command and help message, option `--name` and `--namespace` specified the name and namespace of Alluxio Runtime respectively.

   All the logs will be packed in a package under execution path.
