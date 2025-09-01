package v1alpha1

import "sigs.k8s.io/controller-runtime/pkg/client"

// +kubebuilder:object:generate=false
type StatusManager interface {
	GetStatus() *RuntimeStatus

	SetStatus(RuntimeStatus)
}

// +kubebuilder:object:generate=false
type RuntimeInterface interface {

	// Replicas gets the replicas of runtime worker
	Replicas() int32

	client.Object

	StatusManager
}
