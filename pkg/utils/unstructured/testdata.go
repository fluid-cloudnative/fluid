package unstructured

const stsYaml = `
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: nginx
  labels:
    app: nginx
spec:
  replicas: 1
  serviceName: "none"
  selector: # define how the deployment finds the pods it manages
    matchLabels:
      app: nginx
  template: # define the pods specifications
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - name: nginx
          image: nginx
          volumeMounts:
            - mountPath: /data
              name: hbase-vol
      volumes:
        - name: hbase-vol
          persistentVolumeClaim:
            claimName: shared-data
`

const tfjobYaml = `
apiVersion: "kubeflow.org/v1"
kind: "TFJob"
metadata:
  name: "mnist"
  namespace: kubeflow
  annotations:
   fluid.io/serverless: true
spec:
  cleanPodPolicy: None 
  tfReplicaSpecs:
    Worker:
      replicas: 1 
      restartPolicy: Never
      template:
        spec:
          containers:
            - name: tensorflow
              image: gcr.io/kubeflow-ci/tf-mnist-with-summaries:1.0
              command:
                - "python"
                - "/var/tf_mnist/mnist_with_summaries.py"
                - "--log_dir=/train/logs"
                - "--learning_rate=0.01"
                - "--batch_size=150"
              volumeMounts:
                - mountPath: "/train"
                  name: "training"
          volumes:
            - name: "training"
              persistentVolumeClaim:
                claimName: "tfevent-volume"  
    PS:
      replicas: 1 
      restartPolicy: Never
      template:
        spec:
          containers:
            - name: tensorflow
              image: gcr.io/kubeflow-ci/tf-mnist-with-summaries:1.0
              command:
                - "python"
                - "/var/tf_mnist/mnist_with_summaries.py"
                - "--log_dir=/train/logs"
                - "--learning_rate=0.01"
                - "--batch_size=150"
              volumeMounts:
                - mountPath: "/train"
                  name: "training"
          volumes:
            - name: "training"
              persistentVolumeClaim:
                claimName: "tfevent-volume"
`

const pytorchYaml = `
apiVersion: "kubeflow.org/v1"
kind: "PyTorchJob"
metadata:
  name: "pytorch-dist-mnist-nccl"
spec:
  pytorchReplicaSpecs:
    Master:
      replicas: 1
      restartPolicy: OnFailure
      template:
        metadata:
          annotations:
            sidecar.istio.io/inject: "false"
          labels:
            lyft.com/ml-platform: ""  
        spec:
          containers:
            - name: pytorch
              image: "OUR_AWS_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/lyftlearnhorovod:8678853078c35bf1d003761a070389ca535a5d03"
              command: 
                - python
              args: 
                - "/mnt/user-home/distributed-training-exploration/pytorchjob_distributed_mnist.py"
                - "--backend"
                - "nccl"
                - "--epochs"
                - "2"
              env:
              - name: NCCL_DEBUG
                value: "INFO"
              - name: NCCL_SOCKET_IFNAME
                value: "eth0"
              resources:
                limits:
                  nvidia.com/gpu: 1
              volumeMounts:
              - mountPath: /mnt/user-home
                name: nfs
          volumes:
          - name: nfs
            persistentVolumeClaim:
              claimName: asaha
          tolerations: 
            - key: lyft.net/gpu
              operator: Equal
              value: dedicated
              effect: NoSchedule
    Worker:
      replicas: 1
      restartPolicy: OnFailure
      template:
        metadata:
          annotations:
            sidecar.istio.io/inject: "false"
          labels:
            lyft.com/ml-platform: ""  
        spec:
          containers:
            - name: pytorch
              image: "OUR_AWS_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/lyftlearnhorovod:8678853078c35bf1d003761a070389ca535a5d03"
              command: 
                - python
              args: 
                - "/mnt/user-home/distributed-training-exploration/pytorchjob_distributed_mnist.py"
                - "--backend"
                - "nccl"
                - "--epochs"
                - "2"
              env:
              - name: NCCL_DEBUG
                value: "INFO"
              - name: NCCL_SOCKET_IFNAME
                value: "eth0"
              resources:
                limits:
                  nvidia.com/gpu: 1
              volumeMounts:
              - mountPath: /mnt/user-home
                name: nfs
          volumes:
          - name: nfs
            persistentVolumeClaim:
              claimName: asaha
          tolerations: 
            - key: lyft.net/gpu
              operator: Equal
              value: dedicated
              effect: NoSchedule
`

const argoYaml string = `
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: secret-example-
spec:
  entrypoint: whalesay
  # To access secrets as files, add a volume entry in spec.volumes[] and
  # then in the container template spec, add a mount using volumeMounts.
  volumes:
  - name: my-secret-vol
    secret:
      secretName: my-secret     # name of an existing k8s secret
  templates:
  - name: whalesay
    container:
      image: alpine:3.7
      command: [sh, -c]
      args: ['
        echo "secret from env: $MYSECRETPASSWORD";
        echo "secret from file: test"
      ']
      # To access secrets as environment variables, use the k8s valueFrom and
      # secretKeyRef constructs.
      env:
      - name: MYSECRETPASSWORD  # name of env var
        valueFrom:
          secretKeyRef:
            name: my-secret     # name of an existing k8s secret
            key: mypassword     # 'key' subcomponent of the secret
      volumeMounts:
      - name: my-secret-vol     # mount file containing secret at /secret/mountpath
        mountPath: "/secret/mountpath"
`

const sparkYaml string = `
apiVersion: "sparkoperator.k8s.io/v1beta2"
kind: SparkApplication
metadata:
  name: spark-pi
  namespace: default
spec:
  type: Scala
  mode: cluster
  image: "gcr.io/spark-operator/spark:v3.1.1"
  imagePullPolicy: Always
  mainClass: org.apache.spark.examples.SparkPi
  mainApplicationFile: "local:///opt/spark/examples/jars/spark-examples_2.12-3.1.1.jar"
  sparkVersion: "3.1.1"
  restartPolicy:
    type: Never
  volumes:
    - name: config-vol
      configMap:
        name: dummy-cm
  driver:
    cores: 1
    coreLimit: "1200m"
    memory: "512m"
    labels:
      version: 3.1.1
    serviceAccount: spark
    volumeMounts:
      - name: config-vol
        mountPath: /opt/spark/mycm
  executor:
    cores: 1
    instances: 1
    memory: "512m"
    labels:
      version: 3.1.1
    volumeMounts:
      - name: config-vol
        mountPath: /opt/spark/mycm
`
