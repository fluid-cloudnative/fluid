# 如何使用其他语言（非Go语言）客户端

Fluid 通过 kube-openapi 与 swagger-codegen 可以支持多种语言，包括 Java、Python等。

### 生成方法

```shell
$ cd $GOPATH/src/github.com/fluid-cloudnative
$ hack/sdk/gen-sdk.sh
```

`hack/sdk/gen-sdk.sh` 脚本会在 sdk 目录下自动生成 Python 和 Java client 代码。如果你想使用其他客户端，可以在脚本最下方加入这一行命令：

```shell
java -jar ${SWAGGER_CODEGEN_JAR} generate -i ${SWAGGER_CODEGEN_FILE} -l <language> -o ${JAVA_SDK_OUTPUT_PATH} -c ${SWAGGER_CODEGEN_CONF} --model-package com.github.fluid-cloudnative.fluid
```

### 如何使用

推荐直接使用 io.kubernetes:client-java 的 ApiClient 和 CustomObjectsApi。 只需要导入 com.github.fluid_cloudnative.fluid.* 的Api定义即可。

```java
import io.kubernetes.client.openapi.ApiClient;
import io.kubernetes.client.openapi.apis.CustomObjectsApi;

import com.github.fluid_cloudnative.fluid.*;

public class MyExample {

    // generate this client from a kubeconfig file or something else
    ApiClient apiClient;

    public void createAlluxioRuntime(String namespace, AlluxioRuntime runtime) throws ApiException {
        CustomObjectsApi customObjectsApi = new CustomObjectsApi(apiClient);
        customObjectsApi.createNamespacedCustomObject(
                AlluxioRuntime.group,
                AlluxioRuntime.version,
                namespace,
                AlluxioRuntime.plural,
                runtime,
                "true"
        );
    }

    public AlluxioRuntime getAlluxioRuntime(String namespace, String name) throws Exception {
        CustomObjectsApi customObjectsApi = new CustomObjectsApi(apiClient);
        Object obj = customObjectsApi.getNamespacedCustomObject(
                AlluxioRuntime.group,
                AlluxioRuntime.version,
                namespace,
                AlluxioRuntime.plural,
                name
        );
        Gson gson = new JSON().getGson();
        return gson.fromJson(gson.toJsonTree(obj).getAsJsonObject(), AlluxioRuntime.class);
    }

    /*
    Note that currently ClientJava can only support merge-patch+json
    */
    public void patchAlluxioRuntime(String namespace, String name, String patchBody) throws ApiException {
        CustomObjectsApi customObjectsApi = new CustomObjectsApi(apiClient);
        customObjectsApi.patchNamespacedCustomObject(
                AlluxioRuntime.group,
                AlluxioRuntime.version,
                namespace,
                AlluxioRuntime.plural,
                name,
                patchpatchBody.getBytes()ody
        );
    }
}

```