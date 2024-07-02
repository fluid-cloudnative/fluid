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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func SetupDataOpJobWatcherWithReconciler(mgr ctrl.Manager, options controller.Options, r Controller) (err error) {
	options.Reconciler = r
	c, err := controller.New(r.ControllerName(), mgr, options)
	if err != nil {
		return err
	}

	jobEventHandler := &opJobEventHandler{}
	err = c.Watch(source.Kind(mgr.GetCache(), r.ManagedResource()), &handler.EnqueueRequestForObject{}, predicate.Funcs{
		CreateFunc: jobEventHandler.onCreateFunc(r),
		UpdateFunc: jobEventHandler.onUpdateFunc(r),
		DeleteFunc: jobEventHandler.onDeleteFunc(r),
	})
	if err != nil {
		log.Error(err, "Failed to watch Pod")
		return err
	}
	return
}
