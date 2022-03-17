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

package kubeclient

import (
	"context"

	"k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetNode gets the latest node info
func GetNode(client client.Reader, name string) (node *v1.Node, err error) {
	key := types.NamespacedName{
		Name: name,
	}

	node = &v1.Node{}

	if err = client.Get(context.TODO(), key, node); err != nil {
		return nil, err
	}
	return node, err
}

// IsReady checks if the node is ready
// If the node is ready,it returns True.Otherwise,it returns False.
func IsReady(node v1.Node) (ready bool) {
	ready = true
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady && condition.Status != v1.ConditionTrue {
			ready = false
			break
		}
	}
	return ready
}
