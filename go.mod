module github.com/fluid-cloudnative/fluid

go 1.13

require (
	github.com/agiledragon/gomonkey v2.0.2+incompatible
	github.com/brahma-adshonor/gohook v1.1.9
	github.com/container-storage-interface/spec v1.2.0
	github.com/docker/go-units v0.4.0
	github.com/go-logr/logr v0.1.0
	github.com/go-openapi/spec v0.19.3
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.3.1
	github.com/kubernetes-csi/csi-lib-utils v0.7.0 // indirect
	github.com/kubernetes-csi/drivers v1.0.2
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/pkg/errors v0.8.1
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/cobra v0.0.5
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553
	google.golang.org/grpc v1.26.0
	gopkg.in/yaml.v2 v2.2.8
	k8s.io/api v0.18.5
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.5
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	k8s.io/kubernetes v1.18.5
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
	sigs.k8s.io/controller-runtime v0.3.0
)

replace k8s.io/api => k8s.io/api v0.18.5

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.5

replace k8s.io/apimachinery => k8s.io/apimachinery v0.18.6-rc.0

replace k8s.io/apiserver => k8s.io/apiserver v0.18.5

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.5

replace k8s.io/client-go => k8s.io/client-go v0.18.5

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.5

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.5

replace k8s.io/code-generator => k8s.io/code-generator v0.18.6-rc.0

replace k8s.io/component-base => k8s.io/component-base v0.18.5

replace k8s.io/cri-api => k8s.io/cri-api v0.18.6-rc.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.5

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.5

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.5

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.5

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.5

replace k8s.io/kubectl => k8s.io/kubectl v0.18.5

replace k8s.io/kubelet => k8s.io/kubelet v0.18.5

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.5

replace k8s.io/metrics => k8s.io/metrics v0.18.5

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.5

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.18.5

replace k8s.io/sample-controller => k8s.io/sample-controller v0.18.5

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.6.0
