package fetchpathfrompvc

import (
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/webhook/plugins"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const Name = "FetchPathFromPVC"

var _ plugins.MutatingHandler = &FetchPathFromPVC{}

type FetchPathFromPVC struct {
	client client.Client
	name   string
}

func NewPlugin(c client.Client) plugins.MutatingHandler {
	return &FetchPathFromPVC{
		client: c,
		name:   Name,
	}
}

func (f FetchPathFromPVC) Mutate(pod *corev1.Pod, m map[string]base.RuntimeInfoInterface) (shouldStop bool, err error) {
	//TODO implement me
	panic("implement me")
}

func (f FetchPathFromPVC) GetName() string {
	//TODO implement me
	panic("implement me")
}
