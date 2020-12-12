package jindo

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
)

func (e *JindoEngine) UpdateDatasetStatus(phase datav1alpha1.DatasetPhase) (err error) {
	/*err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
		if err != nil {
			return err
		}
		datasetToUpdate := dataset.DeepCopy()
		datasetToUpdate.Status.Phase = phase
		err = e.Client.Status().Update(context.TODO(), datasetToUpdate)
		if err != nil {
			e.Log.Error(err, "Update dataset")
			return err
		}
		return nil
	})*/
	return nil
}

func (e *JindoEngine) UpdateCacheOfDataset() (err error) {
	return
}

func (e *JindoEngine) BindToDataset() (err error) {
	return e.UpdateDatasetStatus(datav1alpha1.BoundDatasetPhase)
}
