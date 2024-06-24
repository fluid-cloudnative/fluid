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

package watch

import (
	batchv1 "k8s.io/api/batch/v1"
	"strconv"

	"sigs.k8s.io/controller-runtime/pkg/event"
)

// opJobEventHandler represents the handler for data operation jobs.
type opJobEventHandler struct {
}

func (h *opJobEventHandler) onCreateFunc(r Controller) func(e event.CreateEvent) bool {
	return func(e event.CreateEvent) (onCreate bool) {
		// ignore create event
		job, ok := e.Object.(*batchv1.Job)
		if !ok {
			log.Info("job.onCreateFunc Skip", "object", e.Object)
			return false
		}

		if !JobShouldInQueue(job) {
			log.Info("opJobEventHandler.onCreateFunc skip due to shouldRequeue false")
			return false
		}

		log.V(1).Info("opJobEventHandler.onCreateFunc", "name", job.GetName(), "namespace", job.GetNamespace())
		return true
	}
}

func (h *opJobEventHandler) onUpdateFunc(r Controller) func(e event.UpdateEvent) bool {
	return func(e event.UpdateEvent) (needUpdate bool) {
		jobNew, ok := e.ObjectNew.(*batchv1.Job)
		if !ok {
			log.Info("job.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		jobOld, ok := e.ObjectOld.(*batchv1.Job)
		if !ok {
			log.Info("job.onUpdateFunc Skip", "object", e.ObjectNew)
			return needUpdate
		}

		if jobNew.GetResourceVersion() == jobOld.GetResourceVersion() {
			log.V(1).Info("job.onUpdateFunc Skip due to resourceVersion not changed")
			return needUpdate
		}

		// ignore if it's not fluid label job
		if !JobShouldInQueue(jobNew) {
			log.Info("opJobEventHandler.onUpdateFunc skip due to shouldRequeue false")
			return false
		}

		log.Info("opJobEventHandler.onUpdateFunc", "name", jobNew.GetName(), "namespace", jobNew.GetNamespace())
		return true
	}
}

func (h *opJobEventHandler) onDeleteFunc(r Controller) func(e event.DeleteEvent) bool {
	return func(e event.DeleteEvent) bool {
		// ignore delete event
		return false
	}
}

func JobShouldInQueue(job *batchv1.Job) bool {
	if job == nil {
		return false
	}

	// cron data operation does not set dataflow affinity.
	_, exist := job.Labels["cronjob"]
	if exist {
		return false
	}

	// operations with parallel task does not set dataflow affinity.
	value, exist := job.Labels["parallelism"]
	if exist {
		parallelism, err := strconv.Atoi(value)
		if err != nil || parallelism > 1 {
			log.Info("skip as parallelism exist and not 1", "name", job.GetName(), "namespace", job.GetNamespace())
			return false
		}
	}

	return true
}
