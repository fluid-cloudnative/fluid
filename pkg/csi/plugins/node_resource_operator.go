/*
Copyright 2025 The Fluid Authors.

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

package plugins

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NodeAuthorizedClient interface {
	Get(nodeName string) (*corev1.Node, error)
	Patch(node *corev1.Node, patchType types.PatchType, data []byte) error
}

// restrictedNodeClient uses node binding token with validating policy to avoid security problems.
type restrictedNodeClient struct {
	Client client.Client
}

// kubeletNodeClient uses mounted kubelet config to avoid security problems.
type kubeletNodeClient struct {
	Clientset *kubernetes.Clientset
}

func (p *restrictedNodeClient) Get(nodeName string) (*corev1.Node, error) {
	node := &corev1.Node{}
	key := types.NamespacedName{Name: nodeName}
	if err := p.Client.Get(context.TODO(), key, node); err != nil {
		return nil, err
	}
	return node, nil
}

func (p *restrictedNodeClient) Patch(node *corev1.Node, patchType types.PatchType, data []byte) error {
	err := p.Client.Patch(context.TODO(), node, client.RawPatch(patchType, data))
	return err
}

func (p *kubeletNodeClient) Get(nodeName string) (*corev1.Node, error) {
	return p.Clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
}

func (p *kubeletNodeClient) Patch(node *corev1.Node, patchType types.PatchType, data []byte) error {
	_, err := p.Clientset.CoreV1().Nodes().Patch(context.TODO(), node.Name, patchType, data, metav1.PatchOptions{})
	return err
}
