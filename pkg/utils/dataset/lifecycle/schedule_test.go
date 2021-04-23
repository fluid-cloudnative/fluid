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

package lifecycle

import (
	"math/rand"
	"strconv"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestAssignDatasetToNodes(t *testing.T) {

	var nodes []corev1.Node
	pvcMountNodesMap := map[string]int64{}

	fuseSelectLabel := map[string]string{"fuse": "true"}
	fuseNotSelectLabel := map[string]string{"fuse": "false"}

	for i := 1; i <= 100; i++ {
		node := corev1.Node{}
		nodeName := "node" + strconv.Itoa(i)
		node.Name = nodeName
		pvcMountPodsNum := rand.Int63n(5)
		if pvcMountPodsNum != 0 {
			pvcMountNodesMap[nodeName] = pvcMountPodsNum
			node.Labels = fuseSelectLabel
		} else {
			fuseSelect := rand.Intn(2)
			if fuseSelect == 1 {
				node.Labels = fuseSelectLabel
			} else {
				node.Labels = fuseNotSelectLabel
			}
		}
		nodes = append(nodes, node)
	}
	nodes = sortNodesToBeScheduled(nodes, pvcMountNodesMap, fuseSelectLabel)

	for i := 0; i < len(nodes)-1; i++ {
		if nodes[i].Labels["fuse"] == "false" && nodes[i+1].Labels["fuse"] == "true" {
			t.Errorf("the result of sort is not right")
		}

		numFront, found := pvcMountNodesMap[nodes[i].Name]
		if !found {
			numFront = 0
		}
		numBehind, found := pvcMountNodesMap[nodes[i+1].Name]
		if !found {
			numBehind = 0
		}
		if numFront < numBehind {
			t.Errorf("the result of sort is not right")
		}

	}
}
