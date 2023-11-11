/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
*/

package watch

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func SetupAppWatcherWithReconciler(mgr ctrl.Manager, options controller.Options, r Controller) (err error) {
	options.Reconciler = r
	c, err := controller.New(r.ControllerName(), mgr, options)
	if err != nil {
		return err
	}

	podEventHandler := &podEventHandler{}
	err = c.Watch(&source.Kind{Type: r.ManagedResource()}, &handler.EnqueueRequestForObject{}, predicate.Funcs{
		CreateFunc: podEventHandler.onCreateFunc(r),
		UpdateFunc: podEventHandler.onUpdateFunc(r),
		DeleteFunc: podEventHandler.onDeleteFunc(r),
	})
	if err != nil {
		log.Error(err, "Failed to watch Pod")
		return err
	}
	return
}
