package kubelet

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// InitNodeAuthorizedClient initializes node authorized client with kubelet's kube config.
// This is now an available workaround to implement a node-scoped daemonset.
// See discussion https://github.com/kubernetes/enhancements/pull/944#issuecomment-490242290
func InitNodeAuthorizedClient(kubeletKubeConfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeletKubeConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, "fail to build kubelet config")
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "fail to build client-go client from kubelet kubeconfig")
	}

	return client, nil
}
