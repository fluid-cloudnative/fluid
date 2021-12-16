package builder

import (
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Info struct {

	// The input object
	InputObject runtime.Unstructured

	// The output object
	OutputObject runtime.Unstructured
}

func buildInfo(obj runtime.Object) (info *Info, err error) {
	c, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return info, err
	}
	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(c)
	info = &Info{
		InputObject:  u,
		OutputObject: u.DeepCopy(),
	}

	return
}

// Builder provides convenience functions for taking arguments and parameters
// from the command line and converting them to a list of resources to iterate
// over using the Visitor interface.
type Builder struct {
	// The helper
	helper ctrl.Helper

	visitors []Visitor
}

func NewBuilder(obj runtime.Object) *Builder {
	return &Builder{}
}

func (b *Builder) WithHelper(helper ctrl.Helper) *Builder {
	b.helper = helper
	return b
}

func (b *Builder) RegisterHandlers(visitors ...Visitor) *Builder {
	if len(b.visitors) == 0 {
		b.visitors = []Visitor{}
	}
	b.visitors = append(b.visitors, visitors...)
	return b
}

func (b *Builder) Do(obj runtime.Object) (result *Info, err error) {

	info, err := buildInfo(obj)
	if err != nil {
		return info, err
	}

	for _, v := range b.visitors {
		err = v.Visit(info, b.helper, err)
		if err != nil {
			return info, err
		}
	}

	return
}
