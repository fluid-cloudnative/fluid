/*
 Copyright 2022 The Fluid Authors.

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

package webhook

import (
	"context"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/fluid-cloudnative/fluid/pkg/webhook"
	ctrl "sigs.k8s.io/controller-runtime"
)

type WebhookReconciler struct {
	CertBuilder *webhook.CertificateBuilder
	WebhookName string
	CaCert      []byte
}

func (r *WebhookReconciler) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	// patch ca of MutatingWebhookConfiguration
	err := r.CertBuilder.PatchCABundle(r.WebhookName, r.CaCert)
	if err != nil {
		return utils.RequeueAfterInterval(10 * time.Second)
	}
	return utils.NoRequeue()
}
