/*
Copyright 2022 The Kruise Authors.

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

package podreadiness

import (
	"sort"

	"github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/podadapter"
	"github.com/fluid-cloudnative/fluid/pkg/controllers/workload/v1alpha1/utils/util"
	v1 "k8s.io/api/core/v1"

	workloadv1alpha1 "github.com/fluid-cloudnative/fluid/api/workload/v1alpha1"
)

type Interface interface {
	ContainsReadinessGate(pod *v1.Pod) bool
	AddNotReadyKey(pod *v1.Pod, msg Message) error
	RemoveNotReadyKey(pod *v1.Pod, msg Message) error
}

func NewForAdapter(adp podadapter.Adapter) Interface {
	return &commonControl{adp: adp}
}

type commonControl struct {
	adp podadapter.Adapter
}

func (c *commonControl) ContainsReadinessGate(pod *v1.Pod) bool {
	return containsReadinessGate(pod, workloadv1alpha1.KruisePodReadyConditionType)
}

func (c *commonControl) AddNotReadyKey(pod *v1.Pod, msg Message) error {
	return addNotReadyKey(c.adp, pod, msg, workloadv1alpha1.KruisePodReadyConditionType)
}

func (c *commonControl) RemoveNotReadyKey(pod *v1.Pod, msg Message) error {
	return removeNotReadyKey(c.adp, pod, msg, workloadv1alpha1.KruisePodReadyConditionType)
}

type Message struct {
	UserAgent string `json:"userAgent"`
	Key       string `json:"key"`
}

type messageList []Message

func (c messageList) Len() int      { return len(c) }
func (c messageList) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c messageList) Less(i, j int) bool {
	if c[i].UserAgent == c[j].UserAgent {
		return c[i].Key < c[j].Key
	}
	return c[i].UserAgent < c[j].UserAgent
}

func (c messageList) dump() string {
	sort.Sort(c)
	return util.DumpJSON(c)
}
