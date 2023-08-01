package efc

import (
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/dataprocess"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (e *EFCEngine) generateDataProcessValueFile(ctx cruntime.ReconcileRequestContext, object client.Object) (valueFileName string, err error) {
	dataProcess, ok := object.(*datav1alpha1.DataProcess)
	if !ok {
		err = fmt.Errorf("object %v is not of type DataProcess", object)
		return "", err
	}

	targetDataset, err := utils.GetDataset(e.Client, dataProcess.Spec.Dataset.Name, dataProcess.Spec.Dataset.Namespace)
	if err != nil {
		return "", errors.Wrap(err, "failed to get dataset")
	}

	return dataprocess.GenDataProcessValueFile(targetDataset, dataProcess)
}
