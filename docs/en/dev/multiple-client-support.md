## How to user multiple clients other than go

Fluid using kube-openapi and swagger-codegen to support multiple clients, including Java and Python.

Java Client: https://github.com/fluid-cloudnative/fluid-client-java

### how to generate

```shell
$ cd $GOPATH/src/github.com/fluid-cloudnative
$ hack/sdk/gen-sdk.sh
```

`hack/sdk/gen-sdk.sh` script will generate Java and Python client in sdk folder. If you want to generate other client, 
You can modify script by adding these code:

```shell
java -jar ${SWAGGER_CODEGEN_JAR} generate -i ${SWAGGER_CODEGEN_FILE} -l <language> -o ${JAVA_SDK_OUTPUT_PATH} -c ${SWAGGER_CODEGEN_CONF} --model-package com.github.fluid-cloudnative.fluid
```

### How to Use

It is suggested that you should use ApiClient and CustomObjectsApi from io.kubernetes:client-java. 
The only thing you should import from fluid-cloudnative.fluid:fluid-client-java is com.github.fluid_cloudnative.fluid.*

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