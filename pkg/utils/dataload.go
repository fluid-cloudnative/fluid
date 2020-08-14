/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"context"
	"fmt"
	datav1alpha1 "github.com/cloudnativefluid/fluid/api/v1alpha1"
	"github.com/cloudnativefluid/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

/*
* Get DataLoad object with name and namespace.
 */
func GetDataLoad(client client.Client, name, namespace string) (*datav1alpha1.AlluxioDataLoad, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var dataload datav1alpha1.AlluxioDataLoad
	if err := client.Get(context.TODO(), key, &dataload); err != nil {
		return nil, err
	}
	return &dataload, nil
}

/*
* list all DataLoad objects in a specific namespace and return the first one that satisfies a predicate.
* The predicate should be a function which returns a bool given a DataLoad object
 */
func FindDataLoadWithPredicate(c client.Client, namespace string, predFunc func(dl datav1alpha1.AlluxioDataLoad) bool) (dataload *datav1alpha1.AlluxioDataLoad, err error) {
	var dataloadList *datav1alpha1.AlluxioDataLoadList = &datav1alpha1.AlluxioDataLoadList{}
	err = c.List(context.TODO(), dataloadList, &client.ListOptions{Namespace: namespace})
	if err != nil {
		return nil, err
	}
	items := dataloadList.Items
	for _, item := range items {
		if predFunc(item) {
			return &item, nil
		}
	}
	return nil, nil
}

/*
* Generate a new release Name for DataLoad
 */
func NewReleaseName(datasetName string) string {
	return fmt.Sprintf("%s-load-%s", datasetName, RandomAlphaNumberString(common.DATALOAD_SUFFIX_LENGTH))
}

/*
* Return the related job name given a release name.
* A release name should be like <dataset>-load-<random_suffix>,
* and the returned job name will be like <dataset>-loader-job-<random_suffix>
 */
func GetJobNameFromReleaseName(releaseName string) string {
	strs := strings.Split(releaseName, "-load-")
	datasetName := strs[0]
	suffix := strs[len(strs)-1]
	return fmt.Sprintf("%s-loader-job-%s", datasetName, suffix)
}
