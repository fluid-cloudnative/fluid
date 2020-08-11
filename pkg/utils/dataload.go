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

func NewReleaseName(datasetName string) string {
	return fmt.Sprintf("%s-load-%s", datasetName, RandomAlphaNumberString(common.Suffix_length))
}

func GetJobNameFromReleaseName(releaseName string) string {
	strs := strings.Split(releaseName, "-load-")
	datasetName := strs[0]
	suffix := strs[len(strs)-1]
	return fmt.Sprintf("%s-loader-job-%s", datasetName, suffix)
}
