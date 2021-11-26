module github.com/fluid-cloudnative/fluid

go 1.16

replace k8s.io/api => k8s.io/api v0.22.3

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.3

replace k8s.io/apimachinery => k8s.io/apimachinery v0.22.4-rc.0

replace k8s.io/apiserver => k8s.io/apiserver v0.22.3

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.3

replace k8s.io/client-go => k8s.io/client-go v0.22.3

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.22.3

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.22.3

replace k8s.io/code-generator => k8s.io/code-generator v0.22.4-rc.0

replace k8s.io/component-base => k8s.io/component-base v0.22.3

replace k8s.io/component-helpers => k8s.io/component-helpers v0.22.3

replace k8s.io/controller-manager => k8s.io/controller-manager v0.22.3

replace k8s.io/cri-api => k8s.io/cri-api v0.23.0-alpha.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.22.3

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.22.3

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.22.3

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.22.3

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.22.3

replace k8s.io/kubectl => k8s.io/kubectl v0.22.3

replace k8s.io/kubelet => k8s.io/kubelet v0.22.3

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.22.3

replace k8s.io/metrics => k8s.io/metrics v0.22.3

replace k8s.io/mount-utils => k8s.io/mount-utils v0.22.4-rc.0

replace k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.22.3

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.22.3

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.22.3

replace k8s.io/sample-controller => k8s.io/sample-controller v0.22.3

require (
	github.com/agiledragon/gomonkey v2.0.2+incompatible
	github.com/brahma-adshonor/gohook v1.1.9
	github.com/container-storage-interface/spec v1.5.0
	github.com/docker/go-units v0.4.0
	github.com/go-logr/logr v0.4.0
	github.com/go-openapi/spec v0.20.4
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/mock v1.4.4
	github.com/kubernetes-csi/csi-lib-utils v0.10.0 // indirect
	github.com/kubernetes-csi/drivers v1.0.2
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/goconvey v1.6.4
	github.com/spf13/cobra v1.1.3
	go.uber.org/zap v1.19.0
	golang.org/x/net v0.0.0-20210520170846-37e1c6afe023
	google.golang.org/grpc v1.38.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.3
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e
	k8s.io/kubernetes v1.22.3
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a
	sigs.k8s.io/controller-runtime v0.10.3
	sigs.k8s.io/yaml v1.2.0
)
