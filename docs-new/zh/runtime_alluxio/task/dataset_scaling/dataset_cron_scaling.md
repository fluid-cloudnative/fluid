# 示例 - Dataset 定时弹性扩缩容

## 背景

根据数据缓存量比例触发自动的数据缓存能力弹性扩缩容，它有非常多的优势，但是有一个缺陷，就是需要根据资源压力计算出合理的值后调整，这就存在一定的程度滞后性。而针对此，Fluid结合CronHPA提供了数据访问加速能力的定时扩缩方案。

首先了解一下具体场景。以离线模型训练场景为例,其中主要有两大场景：

- 实验性模型训练,用户没有固定的模型训练时间和更新频率,一般是一次性实验,用于新模型的测试,或是为已有模型选择新的训练样本,调参等。  
- 周期性迭代训练,即用户按照固定的频率进行模型训练及模型发布上线,比如日更模型,周更模型,月更模型等。

在实际的生产环境中,集群60%以上时间用于周期性迭代训练,因此如何提高周期性训练的效率, 降低周期性训练的成本是在建设周期性迭代模型链路时,优先考虑的。我们通过Fluid提升了计算数据访问的效率，进而优化了集群计算资源的效率和运算速度。但是在实践中，我们发现这里还有些优化的空间。比如通过观察可以发现了以下问题:

1)以日更模型举例,真正的模型训练需要等待上游数据的预处理,而这段时间的Fluid缓存资源一般是浪费的;

下图是一个典型的日更训练任务使用的fluid节点的io变化曲线。新的日更模型已经在晚8点前上线,而凌晨到早晨6点前则是在进行训练数据的预处理,因此在这10个小时内无任何使用这个特定dataset的训练任务,Fluid占用的资源是被浪费的。

![dataset.jpg](https://ucc.alicdn.com/pic/developer-ecology/4ac833ce69f14dc3a6582afad7a83fcd.jpg)


2)周期性模型使用的训练数据仅在这个训练周期内使用频率较高,超过这个训练周期后,数据的使用频率较低或者基本不会再被使用;

3)在训练任务密集的白天,多个Fluid的数据集会通过自动扩容进行资源的抢占,这些数据集在使用频率较低时仍会占用大量的资源,导致需要进行扩容的数据集没有资源可扩,严重影响了训练效率


针对上述问题,Fluid提供了数据缓存的弹性伸缩能力, 甚至可将缓存能力缩减为0。而且动态调整缓存容量变得非常简单，只需要[修改Runtime的replicas](./dataset_scaling.md)就可以完成数据缓存的扩缩容。如果能够配合定时弹性伸缩的控制器，就可以实现数据缓存的按需使用，充分发挥资源的有效性。恰好我们发现开源社区的kubernetes-cronhpa-controller可以很好的解决拥有周期性资源画像的负载弹性，结合底层的cluster-autoscaler可以降低大量的资源成本。目前kubernetes-cronhpa-controller已经开源有两年了，并且在许多真实场景下打磨成熟。具体实现在[开源代码仓库](https://github.com/AliyunContainerService/kubernetes-cronhpa-controller)

在本文中，我们就像您展示如何实现数据访问加速能力的定时扩缩。

## 实践操作


1.安装jq工具方便解析json，在本例子中我们使用操作系统是centos，可以通过yum安装jq

```
yum install -y jq
```

2.下载、安装Fluid最新版

```
git clone https://github.com/fluid-cloudnative/fluid.git
cd fluid/charts
kubectl create ns fluid-system
helm install fluid fluid
```

3.部署或配置 kubernetes-cronhpa-controller

```shell
$ cd -
$ kubectl apply -f fluid/integration/cronhpa
```

4.验证kubernetes-cronhpa-controller安装状态

```shell
$ kubectl get deploy kubernetes-cronhpa-controller -n kube-system
NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
kubernetes-cronhpa-controller   1/1     1            1           6d5h
```

5.提交测试使用的Dataset

```shell
$ cat<<EOF >dataset.yaml
apiVersion: data.fluid.io/v1alpha1
kind: Dataset
metadata:
  name: spark
spec:
  mounts:
    - mountPoint: https://mirrors.bit.edu.cn/apache/spark/
      name: spark
---
apiVersion: data.fluid.io/v1alpha1
kind: AlluxioRuntime
metadata:
  name: spark
spec:
  replicas: 1
  tieredstore:
    levels:
      - mediumtype: MEM
        path: /dev/shm
        quota: 1Gi
        high: "0.99"
        low: "0.7"
EOF
$ kubectl create -f dataset.yaml
dataset.data.fluid.io/spark created
alluxioruntime.data.fluid.io/spark created
```

6.查看这个Dataset是否处于可用状态, 可以看到该数据集的总量为2.71GiB， 目前Fluid提供的缓存节点数为1，可以提供的最大缓存能力为1GiB。此时数据量是无法满足全量数据缓存的需求。

```shell
$ kubectl get dataset
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
spark   2.71GiB          0.00B    1.00GiB          0.0%                Bound   7m38s
```

此时支持该dataset的alluxioruntime

```shell
kubectl get alluxioruntimes.data.fluid.io -owide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE   AGE
spark   1               1                 Ready          1               1                 Ready          0             0               Ready        104s
```

7.创建cronHPA任务

```shell
$ cat<<EOF > hpa.yaml
apiVersion: autoscaling.alibabacloud.com/v1beta1
kind: CronHorizontalPodAutoscaler
metadata:
  name: spark
  namespace: default
spec:
   scaleTargetRef:
      apiVersion: data.fluid.io/v1alpha1
      kind: AlluxioRuntime
      name: spark
   excludeDates:
   # exclude May 1st
   - "* * * 1 5 *"
   jobs:
   - name: "scale-down"
     schedule: "0 0 8 ? * 1-6"
     targetSize: 0
   - name: "scale-up"
     schedule: "0 30 21 ? * 1-5"
     targetSize: 3
EOF
```

首先，我们解读一下从样例配置，这里主要有三部分：

- **伸缩对象**：其中`scaleTargetRef`字段描述伸缩的对象，这里的操作对象为`AlluxioRuntime`

- **日期过滤**： 不同类型应用画像也不尽相同，比如线应用类型的，还有离线任务类型的，他们的资源使用画像也是各不相同。有些工作任务在法定节假日就是波谷，因此可以提供关闭伸缩规则的时间。比如在本例子中就规定5月1日规则不生效。

- **伸缩规则**： 该规则列表中，可以定义多个规则。每个规则由四部分组成，分别是同一个CronHPA的jobs列表中唯一的`name`；定义任务执行时间规则的`schedule`,它的规则和crontab类似；`targetSize`为到调度时间时，扩缩容的目标数目；`runOnce`如果数值为true，则代表该任务仅执行一次。
更多细节可以查看[官方文档](https://help.aliyun.com/document_detail/151557.html)

在本例子中，扩容任务会在每周的周一到周五晚上21：30执行，扩容目标为3；缩容任务会在每周的周一到周六的早上8点执行，目标是0。而五月一日该定时任务会暂停。


8.时隔一周之后，我们查看一下该CronHPA任务执行效果。首先查看该CronHPA的状态, 可以看到`scale-up`和`scale-down`的任务都已经完成

```shell
kubectl describe cronhorizontalpodautoscalers.autoscaling.alibabacloud.com spark
Name:         spark
Namespace:    default
Annotations:  <none>
API Version:  autoscaling.alibabacloud.com/v1beta1
Kind:         CronHorizontalPodAutoscaler
Metadata:
  Creation Timestamp:  2021-04-12T07:01:54Z
  Generation:          12
  Resource Version:  4922900
  Self Link:         /apis/autoscaling.alibabacloud.com/v1beta1/namespaces/default/cronhorizontalpodautoscalers/spark
  UID:               a156cfb9-e491-43a5-9959-494395ac350b
Spec:
  Exclude Dates:
    * * * 1 5 *
  Jobs:
    Name:         scale-down
    Schedule:     0 0 8 ? * 1-6
    Target Size:  0
    Name:         scale-up
    Schedule:     0 30 21 ? * 1-5
    Target Size:  3
  Scale Target Ref:
    API Version:  data.fluid.io/v1alpha1
    Kind:         AlluxioRuntime
    Name:         spark
Status:
  Conditions:
    Job Id:           ff0eb79b-8e44-4bfc-9872-8cb75f07c656
    Last Probe Time:  2021-04-16T13:30:00Z
    Message:          cron hpa job scale-up executed successfully. current replicas:0, desired replicas:3.
    Name:             scale-up
    Run Once:         false
    Schedule:         0 30 21 ? * 1-5
    State:            Succeed
    Target Size:      3
    Job Id:           6f3930ee-241d-4549-bf22-93b017cef4a4
    Last Probe Time:  2021-04-17T00:00:00Z
    Message:          cron hpa job scale-down executed successfully. current replicas:3, desired replicas:0.
    Name:             scale-down
    Run Once:         false
    Schedule:         0 0 8 ? * 1-6
    State:            Succeed
    Target Size:      0
  Exclude Dates:
    * * * 1 5 *
  Scale Target Ref:
    API Version:  data.fluid.io/v1alpha1
    Kind:         AlluxioRuntime
    Name:         spark
Events:           <none>
```

9.由于我们的查看时段是选在白天，此时缩容已经完成。此时我们期望看到的结果是缓存能力为0，执行一下查询命令进行确认确实缓存能力为0。说明此时缓存能力已经由创建时刻的1缩容到0，定时扩缩容任务已经生效。

```shell
$ kubectl get dataset
NAME    UFS TOTAL SIZE   CACHED   CACHE CAPACITY   CACHED PERCENTAGE   PHASE   AGE
spark   2.71GiB          0.00B    0.00B            0.0%                Bound   6d19h
```

进一步，我们查看一下此时alluxio runtime的replicas数量，也为0。可以看到kubernetes-cronhpa-controller将缓存引擎的副本数从初始时的1缩容到0.

```shell
$ kubectl get alluxioruntimes.data.fluid.io -owide
NAME    READY MASTERS   DESIRED MASTERS   MASTER PHASE   READY WORKERS   DESIRED WORKERS   WORKER PHASE   READY FUSES   DESIRED FUSES   FUSE PHASE  AGE
spark   1               1                 Ready          0               0                 Ready          0             0               6d19h
```

10.如果希望查询定时扩容历史记录还可以通过CronHPA提供运维页面, 可以进一步了解该任务的上次和下次扩缩容时间

首先查询该运维页面的访问端口, 可以看到该service的端口为80

```shell
$ kubectl get svc -n kube-system kubernetes-cronhpa-controller
NAME                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
kubernetes-cronhpa-controller   ClusterIP   172.21.11.139   <none>        80/TCP    30s
```

此时，在本地可以通过执行`kubectl proxy --port=8080`启动proxy模式，再在本地浏览器中输入http://localhost:8080/api/v1/namespaces/kube-system/services/kubernetes-cronhpa-controller/proxy/ 即可访问运维管理页面。

![cronhpa.png](https://ucc.alicdn.com/pic/developer-ecology/256c66f593d24b27944d40ba755c1ca4.png)



## 总结

Fluid结合CronHPA提供了定时扩缩容的能力，结合应用自身使用数据的特点，实现了数据缓存的按时扩缩容，充分的利用了集群计算和存储资源加速应用的数据访问性能。现阶段使用自动扩容+定时缩容可以最大化的使Fluid存储在集群内变成一种可控的弹性存储资源。下一步我们将对缩容进行时的数据迁移和平衡(rebalance)进行支持,保证缩容时数据的动态平衡。
