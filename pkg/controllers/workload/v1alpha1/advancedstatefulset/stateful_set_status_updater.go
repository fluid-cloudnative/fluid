/*
Copyright 2026 The Fluid Authors.
Copyright 2019 The Kruise Authors.
Copyright 2017 The Kubernetes Authors.

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

package advancedstatefulset

import (
	"context"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	workloadv1alpha1 "github.com/fluid-cloudnative/fluid/api/workload/v1alpha1"
)

// StatusUpdaterInterface is an interface used to update the AdvancedStatefulSetStatus associated with a AdvancedStatefulSet.
// For any use other than testing, clients should create an instance using NewRealStatefulSetStatusUpdater.
type StatusUpdaterInterface interface {
	// UpdateStatefulSetStatus sets the set's Status to status. Implementations are required to retry on conflicts,
	// but fail on other errors. If the returned error is nil set's Status has been successfully set to status.
	UpdateStatefulSetStatus(ctx context.Context, set *workloadv1alpha1.AdvancedStatefulSet, status *workloadv1alpha1.AdvancedStatefulSetStatus) error
}

// NewRealStatefulSetStatusUpdater returns a StatusUpdaterInterface that updates the Status of a AdvancedStatefulSet,
// using the supplied controller-runtime client.
func NewRealStatefulSetStatusUpdater(c client.Client) StatusUpdaterInterface {
	return &realStatefulSetStatusUpdater{client: c}
}

type realStatefulSetStatusUpdater struct {
	client client.Client
}

func (ssu *realStatefulSetStatusUpdater) UpdateStatefulSetStatus(
	ctx context.Context,
	set *workloadv1alpha1.AdvancedStatefulSet,
	status *workloadv1alpha1.AdvancedStatefulSetStatus) error {
	// don't wait due to limited number of clients, but backoff after the default number of steps
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Fetch latest version to apply status update
		fresh := &workloadv1alpha1.AdvancedStatefulSet{}
		if err := ssu.client.Get(ctx, types.NamespacedName{Namespace: set.Namespace, Name: set.Name}, fresh); err != nil {
			return err
		}
		fresh.Status = *status
		return ssu.client.Status().Update(ctx, fresh)
	})
}

var _ StatusUpdaterInterface = &realStatefulSetStatusUpdater{}
