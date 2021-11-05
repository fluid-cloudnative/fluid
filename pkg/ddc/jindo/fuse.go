/*
Copyright 2021 The Fluid Authors.

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

package jindo

import (
	"context"
	"fmt"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// cleanupFuse will cleanup Fuse
func (e *JindoEngine) cleanupFuse() (count int, err error) {

	var (
		nodeList     = &corev1.NodeList{}
		fuseLabelKey = common.LabelAnnotationFusePrefix + e.namespace + "-" + e.name
	)

	labelNames := []string{fuseLabelKey}
	e.Log.Info("check node labels", "labelNames", labelNames)
	fuseLabelSelector, err := labels.Parse(fmt.Sprintf("%s=true", fuseLabelKey))
	if err != nil {
		return
	}

	err = e.List(context.TODO(), nodeList, &client.ListOptions{
		LabelSelector: fuseLabelSelector,
	})
	if err != nil {
		return count, err
	}

	nodes := nodeList.Items
	if len(nodes) == 0 {
		e.Log.Info("No node with fuse label need to be delete")
		return
	} else {
		e.Log.Info("Try to clean the fuse label for nodes", "len", len(nodes))
	}

	var labelsToModify common.LabelsToModify
	labelsToModify.Delete(fuseLabelKey)

	for _, node := range nodes {
		_, err = utils.ChangeNodeLabelWithPatchMode(e.Client, &node, labelsToModify)
		if err != nil {
			e.Log.Error(err, "Error when patching labels on node", "nodeName", node.Name)
			return count, errors.Wrapf(err, "NodeStageVolume: error when patching labels on node %s", node.Name)
		}
		count++
	}

	return
}
