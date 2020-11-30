package jindo

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

func (e *JindoEngine) CheckMasterReady() (ready bool, err error) {
	return true, nil
}

// ShouldSetupMaster checks if we need setup the master
func (e *JindoEngine) ShouldSetupMaster() (should bool, err error) {
	runtime, err := e.getRuntime()
	if err != nil {
		return
	}

	switch runtime.Status.MasterPhase {
	case datav1alpha1.RuntimePhaseNone:
		should = true
	default:
		should = false
	}

	return
}

// SetupMaster setups the master and updates the status
// It will print the information in the Debug window according to the Master status
// It return any cache error encountered
func (e *JindoEngine) SetupMaster() (err error) {

	// Setup the Jindo cluster
	master, err := e.getMasterStatefulset(e.name+"-jindofs-master", e.namespace)
	if err != nil && apierrs.IsNotFound(err) {
		//1. Is not found error
		e.Log.V(1).Info("SetupMaster", "master", e.name+"-master")
		return e.setupMasterInernal()
	} else if err != nil {
		//2. Other errors
		return
	} else {
		//3.The master has been set up
		e.Log.V(1).Info("The master has been set.", "replicas", master.Status.ReadyReplicas)
	}

	return nil
}
