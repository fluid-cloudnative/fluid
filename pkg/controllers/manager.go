package controllers

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

// NewCacheClientBypassSecrets creates a client querying kubernetes resources with cache(informers) except for Secrets.
// Secret is an exception for that we tries to trade performance for higher security(e.g. less rbac verbs on Secrets).
func NewCacheClientBypassSecrets(cache cache.Cache, config *rest.Config, options client.Options, uncachedObjects ...client.Object) (client.Client, error) {
	return cluster.DefaultNewClient(cache, config, options, &corev1.Secret{})
}
