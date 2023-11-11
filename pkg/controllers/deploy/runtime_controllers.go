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

package deploy

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"github.com/pkg/errors"
	"reflect"
	"strconv"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	utilpointer "k8s.io/utils/pointer"
)

type CheckFunc func(client.Client, types.NamespacedName) (bool, error)

var precheckFuncs map[string]CheckFunc

func SetPrecheckFunc(checks map[string]CheckFunc) {
	precheckFuncs = checks
}

func ScaleoutRuntimeContollerOnDemand(c client.Client, datasetKey types.NamespacedName, log logr.Logger) (
	controllerName string, scaleout bool, err error) {

	for myControllerName, checkRuntime := range precheckFuncs {
		match, err := checkRuntime(c, datasetKey)
		if err != nil {
			return controllerName, scaleout, err
		}

		if match {
			namespace, err := utils.GetEnvByKey(common.MyPodNamespace)
			if err != nil {
				return controllerName, scaleout, errors.Wrapf(err, "get namespace from env failed, env key:%s", common.MyPodNamespace)
			}
			if namespace == "" {
				namespace = common.NamespaceFluidSystem
			}
			scaleout, err = scaleoutDeploymentIfNeeded(c, types.NamespacedName{
				Namespace: namespace,
				Name:      myControllerName,
			}, log)
			if err != nil {
				return controllerName, scaleout, err
			}
			// if it's match, the skip checking other runtime controller
			return myControllerName, scaleout, nil
		}

	}

	// no matched controller
	return controllerName, scaleout, fmt.Errorf("no matched controller for dataset %s", datasetKey)
}

// scaleoutDeploymentIfNeeded scales out runtime controller deployments if the current replica of it is 0.
func scaleoutDeploymentIfNeeded(c client.Client, key types.NamespacedName, log logr.Logger) (scale bool, err error) {
	err = retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		deploy := &appsv1.Deployment{}
		err = c.Get(context.TODO(), key, deploy)
		if err != nil {
			return err
		}
		deployToUpdate := deploy.DeepCopy()
		// scale out at least 1
		if *deployToUpdate.Spec.Replicas == 0 {
			replicasStr, ok := deployToUpdate.Annotations[common.RuntimeControllerReplicas]
			var replicas int32 = 0
			if ok {
				replicasInt64, _ := strconv.ParseInt(replicasStr, 10, 32)
				replicas = int32(replicasInt64)
			}
			if replicas <= 1 {
				replicas = 1
			}
			deployToUpdate.Spec.Replicas = utilpointer.Int32(replicas)
			scale = true
		} else {
			log.V(1).Info("No need to scale out runtime controller, skip", "key", key)
			return nil
		}

		if !reflect.DeepEqual(deploy, deployToUpdate) {
			err = c.Update(context.TODO(), deployToUpdate)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Error(err, "Failed to scale deployment", "key", key)
	}

	return
}
