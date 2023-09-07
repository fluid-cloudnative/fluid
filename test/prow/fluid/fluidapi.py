API_VERSION = "data.fluid.io/v1alpha1"
webufs = "https://mirrors.bit.edu.cn/apache/zookeeper/stable/"

oss_bucket = "oss://fluid-e2e"
oss_endpoint = "oss-cn-hongkong-internal.aliyuncs.com"

minio_svc = "minio"


class K8sObject():
    def __init__(self):
        self.resource = {}

    def set_kind(self, api_version, kind):
        self.resource["apiVersion"] = api_version
        self.resource["kind"] = kind

        return self

    def set_namespaced_name(self, namespace, name):
        if "metadata" not in self.resource:
            self.resource["metadata"] = {}

        self.resource["metadata"]["namespace"] = namespace
        self.resource["metadata"]["name"] = name

        return self
    
    def set_annotation(self, key, value):
        if "metadata" not in self.resource:
            self.resource["metadata"] = {}
        
        if "annotations" not in self.resource["metadata"]:
            self.resource["metadata"]["annotations"] = {}
        
        self.resource["metadata"]["annotations"][key] = value

    def dump(self):
        return self.resource


class Mount():
    def __init__(self):
        self.str = {}

    def set_mount_info(self, name, mount_point, path=""):
        self.str["mountPoint"] = mount_point
        self.str["name"] = name
        self.str["path"] = path

    def add_options(self, key, value):
        if "options" not in self.str:
            self.str["options"] = {}

        self.str["options"][key] = value

    def add_encrypt_options(self, key, secretName, secretKey):
        if "encryptOptions" not in self.str:
            self.str["encryptOptions"] = []

        self.str["encryptOptions"].append(
            {
                "name": key,
                "valueFrom": {
                    "secretKeyRef": {
                        "name": secretName,
                        "key": secretKey
                    }
                }
            }
        )

    def dump(self):
        return self.str


class Dataset(K8sObject):
    def __init__(self, name, namespace="default"):
        super().__init__()
        self.set_kind("data.fluid.io/v1alpha1", "Dataset")
        self.set_namespaced_name(namespace, name)

    def add_mount(self, mount):
        if "spec" not in self.resource:
            self.resource["spec"] = {}
        if "mounts" not in self.resource["spec"]:
            self.resource["spec"]["mounts"] = []

        self.resource["spec"]["mounts"].append(mount)
        return self

    def set_placement(self, placement):
        if "spec" not in self.resource:
            self.resource["spec"] = {}
        self.resource["spec"]["placement"] = placement
        return self

    def set_access_mode(self, mode):
        if "spec" not in self.resource:
            self.resource["spec"] = {}
        if "accessModes" not in self.resource["spec"]:
            self.resource["spec"]["accessModes"] = []

        self.resource["spec"]["accessModes"].append(mode)
        return self
    
    def set_node_affinity(self, key, value):
        if "spec" not in self.resource:
            self.resource["spec"] = {}
        if "nodeAffinity" not in self.resource["spec"]:
            self.resource["spec"]["nodeAffinity"] = {
                "required": {
                    "nodeSelectorTerms": [{
                        "matchExpressions": [{
                            "key": key,
                            "operator": "In",
                            "values": [value]
                        }]
                    }]
                }
            }
        
        return self



class Runtime(K8sObject):
    def __init__(self, kind, name, namespace="default"):
        super().__init__()
        self.set_kind("data.fluid.io/v1alpha1", kind)
        self.set_namespaced_name(namespace, name)

    def set_replicas(self, replica_num):
        if "spec" not in self.resource:
            self.resource["spec"] = {}

        self.resource["spec"]["replicas"] = replica_num
        return self

    def set_tieredstore(self, mediumtype, path, quota="", quota_list="", high="0.99", low="0.99"):
        if "spec" not in self.resource:
            self.resource["spec"] = {}

        self.resource["spec"]["tieredstore"] = {
            "levels": [{
                "mediumtype": mediumtype,
                "path": path,
                "high": high,
                "low": low
            }]
        }

        if len(quota) != 0:
            self.resource["spec"]["tieredstore"]["levels"][0]["quota"] = quota
        if len(quota_list) != 0:
            self.resource["spec"]["tieredstore"]["levels"][0]["quota"] = quota_list
        
        return self


class DataLoad(K8sObject):
    def __init__(self, name, namespace="default"):
        super().__init__()
        self.set_kind(API_VERSION, "DataLoad")
        self.set_namespaced_name(namespace, name)

    def set_target_dataset(self, dataset_name, dataset_namespace="default"):
        if "spec" not in self.resource:
            self.resource["spec"] = {}

        if "dataset" not in self.resource["spec"]:
            self.resource["spec"]["dataset"] = {}

        self.resource["spec"]["dataset"]["name"] = dataset_name
        self.resource["spec"]["dataset"]["namespace"] = dataset_namespace

        return self

    def set_load_metadata(self, should_load):
        if "spec" not in self.resource:
            self.resource["spec"] = {}

        self.resource["spec"]["loadMetadata"] = should_load

        return self

    def set_cron(self, schedule):
        if "spec" not in self.resource:
            self.resource["spec"] = {}

        self.resource["spec"]["policy"] = "Cron"
        self.resource["spec"]["schedule"] = schedule

        return self


def assemble_dataset(testcase):
    if testcase == "alluxio-webufs":
        mount = Mount()
        mount.set_mount_info("zookeeper", webufs)

        dataset = Dataset("hbase")
        dataset.add_mount(mount.dump())

        return dataset

    elif testcase == "alluxio-oss":
        mount = Mount()
        mount.set_mount_info("demo", oss_bucket)
        mount.add_options("fs.oss.endpoint", oss_endpoint)
        mount.add_encrypt_options("fs.oss.accessKeyId", "access-key", "fs.oss.accessKeyId")
        mount.add_encrypt_options("fs.oss.accessKeySecret", "access-key", "fs.oss.accessKeySecret")

        dataset = Dataset("alluxio-demo-dataset")
        dataset.add_mount(mount.dump())

        return dataset

    elif testcase == "jindo-oss":
        mount = Mount()
        mount.set_mount_info("demo", oss_bucket, "/")
        mount.add_options("fs.oss.endpoint", oss_endpoint)
        mount.add_encrypt_options("fs.oss.accessKeyId", "access-key", "fs.oss.accessKeyId")
        mount.add_encrypt_options("fs.oss.accessKeySecret", "access-key", "fs.oss.accessKeySecret")

        dataset = Dataset("demo-dataset")
        dataset.add_mount(mount.dump())

        return dataset

    elif testcase == "juicefs-minio":
        mount = Mount()
        mount.set_mount_info("juicefs-community", "juicefs:///")
        mount.add_options("bucket", "http://%s:9000/minio/test" % minio_svc)
        mount.add_options("storage", "minio")
        mount.add_encrypt_options("metaurl", "jfs-secret", "metaurl")
        mount.add_encrypt_options("access-key", "jfs-secret", "accesskey")
        mount.add_encrypt_options("secret-key", "jfs-secret", "secretkey")

        dataset = Dataset("jfsdemo")
        dataset.add_mount(mount.dump())
        dataset.set_access_mode(mode="ReadWriteMany")

        return dataset


def assemble_runtime(testcase):
    if testcase == "alluxio-webufs":
        return __assemble_runtime_by_kind("alluxio", "hbase")
    elif testcase == "alluxio-oss":
        return __assemble_runtime_by_kind("alluxio", "alluxio-demo-dataset")
    elif testcase == "jindo-oss":
        return __assemble_runtime_by_kind("jindo", "demo-dataset")
    elif testcase == "juicefs-minio":
        return __assemble_runtime_by_kind("juicefs", "jfsdemo")


def __assemble_runtime_by_kind(runtime_kind, name):
    if runtime_kind == "alluxio":
        runtime = Runtime("AlluxioRuntime", name)
        runtime.set_replicas(1)
        runtime.set_tieredstore("MEM", "/dev/shm", "4Gi")

        return runtime
    elif runtime_kind == "jindo":
        runtime = Runtime("JindoRuntime", name)
        runtime.set_replicas(1)
        runtime.set_tieredstore("MEM", "/dev/shm", "15Gi")

        return runtime

    elif runtime_kind == "juicefs":
        runtime = Runtime("JuiceFSRuntime", name)
        runtime.set_replicas(1)
        runtime.set_tieredstore("MEM", "/dev/shm/cache1:/dev/shm/cache2", "4Gi", high="", low="0.1")

        return runtime
