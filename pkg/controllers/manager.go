package controllers

import (
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

// NewFluidControllerClient creates client.Client according to the HELM_DRIVER env variable. It returns the default client when setting HELM_DRIVER=true,
// meaning users explicitly grant secret permissions to Fluid controllers. Otherwise, it returns a specific client.Client that utilizes informers as cache
// except for Secrets.
func NewFluidControllerClient(cache cache.Cache, config *rest.Config, options client.Options, uncachedObjects ...client.Object) (client.Client, error) {
	if driver, exist := os.LookupEnv("HELM_DRIVER"); exist && driver == "secret" {
		return cluster.DefaultNewClient(cache, config, options, uncachedObjects...)
	}

	return NewCacheClientBypassSecrets(cache, config, options, uncachedObjects...)
}

// NewCacheClientBypassSecrets creates a client querying kubernetes resources with cache(informers) except for Secrets.
// Secret is an exception for that we aim to trade performance for higher security(e.g. less rbac verbs on Secrets).
func NewCacheClientBypassSecrets(cache cache.Cache, config *rest.Config, options client.Options, uncachedObjects ...client.Object) (client.Client, error) {
	return cluster.DefaultNewClient(cache, config, options, append(uncachedObjects, &corev1.Secret{})...)
}

func GetConfigOrDieWithQPSAndBurst(qps int, burst int) *rest.Config {
	cfg := ctrl.GetConfigOrDie()
	if qps > 0 {
		cfg.QPS = float32(qps)
	}

	if burst > 0 {
		cfg.Burst = burst
	}

	return cfg
}
