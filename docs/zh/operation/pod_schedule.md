# 示例 - 通过Webhook 机制优化Pod 调度

Fluid结合根据数据集排布的Pod调度策略，通过webhook机制将调度信息注入到Pod可以实现以下功能：

1.支持K8s原生调度器,以及Volcano, Yunikorn等实现 Pod 数据亲和性调度
2.在全局Fuse的模式下，将Pod优先调度到有数据缓存能力的节点
3.当Pod不使用数据集时，可以尽量避免调度到有缓存的节点

## 前提条件

您使用的k8s版本需要支持 admissionregistration.k8s.io/v1beta1（ Kubernetes version > 1.14 )

## 安装

1.创建命名空间
```shell
kubectl create ns fluid-system
```
2.下载 fluid-0.6.0.tgz
3.使用 Helm 安装 Fluid

```shell
helm install --set webhook.enabled=true fluid fluid-0.6.0.tgz
```
## 配置

**为namespace添加标签**

为namespace添加标签fluid.io/enable-injection后，可以开启此namespace下Pod的调度优化功能

```bash
$ kubectl label namespace default fluid.io/enable-injection=true
```

如果该命名空间下的某些Pod，您不希望开启调度优化功能，只需为Pod打上标签fluid.io/enable-injection=false

例如，使用yaml文件方式创建一个nginx Pod时，应对yaml文件做如下修改：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
  labels:
    fluid.io/enable-injection: false
```

## 具体使用方式

针对于Dataset中Fuse不同的部署模式，Fluid调度的Pod都可以提供相应的支持，具体的用例请参考以下文档：

- [默认部署模式](pod_schedule_default.md)
- [全局部署模式](pod_schedule_global.md)
