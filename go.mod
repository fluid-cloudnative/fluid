module github.com/fluid-cloudnative/fluid

go 1.16

replace k8s.io/api => k8s.io/api v0.23.0

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.12

replace k8s.io/apimachinery => k8s.io/apimachinery v0.23.0

replace k8s.io/apiserver => k8s.io/apiserver v0.20.12

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.12

replace k8s.io/client-go => k8s.io/client-go v0.23.0

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.12

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.12

replace k8s.io/code-generator => k8s.io/code-generator v0.20.13-rc.0

replace k8s.io/component-base => k8s.io/component-base v0.20.12

replace k8s.io/component-helpers => k8s.io/component-helpers v0.20.12

replace k8s.io/controller-manager => k8s.io/controller-manager v0.20.12

replace k8s.io/cri-api => k8s.io/cri-api v0.20.14-rc.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.12

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.12

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.12

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.12

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.12

replace k8s.io/kubectl => k8s.io/kubectl v0.20.12

replace k8s.io/kubelet => k8s.io/kubelet v0.20.12

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.12

replace k8s.io/metrics => k8s.io/metrics v0.20.12

replace k8s.io/mount-utils => k8s.io/mount-utils v0.20.13-rc.0

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.12

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.20.12

replace k8s.io/sample-controller => k8s.io/sample-controller v0.20.12

require (
	github.com/agiledragon/gomonkey v2.0.2+incompatible
	github.com/brahma-adshonor/gohook v1.1.9
	github.com/container-storage-interface/spec v1.2.0
	github.com/docker/go-units v0.4.0
	github.com/go-logr/logr v1.2.0
	github.com/go-openapi/spec v0.19.3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.5.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/kubernetes-csi/csi-lib-utils v0.9.1 // indirect
	github.com/kubernetes-csi/drivers v1.0.2
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	go.uber.org/tools v0.0.0-20190618225709-2cfd321de3ee // indirect
	go.uber.org/zap v1.19.1
	golang.org/x/net v0.0.0-20210825183410-e898025ed96a
	google.golang.org/grpc v1.38.0
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.23.0
	k8s.io/apimachinery v0.23.0
	k8s.io/client-go v0.23.0
	k8s.io/component-base v0.23.0
	k8s.io/component-helpers v0.20.12
	k8s.io/klog/v2 v2.30.0
	k8s.io/kube-openapi v0.0.0-20211115234752-e816edb12b65
	k8s.io/kubernetes v1.20.12
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b
	sigs.k8s.io/controller-runtime v0.11.1
	sigs.k8s.io/yaml v1.3.0
)
