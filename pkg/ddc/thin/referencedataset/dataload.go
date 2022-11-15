package referencedataset

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
)

func (t *ReferenceDatasetEngine) LoadData(ctx cruntime.ReconcileRequestContext, targetDataload datav1alpha1.DataLoad) (err error) {
	//TODO implement me
	return nil
}

func (t *ReferenceDatasetEngine) CheckRuntimeReady() (ready bool) {
	//TODO implement me
	return true
}

func (t *ReferenceDatasetEngine) CheckExistenceOfPath(targetDataload datav1alpha1.DataLoad) (notExist bool, err error) {
	//TODO implement me
	return true, nil
}
