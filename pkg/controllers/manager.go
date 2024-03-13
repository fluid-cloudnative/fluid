package controllers

import (
	"os"
	"time"

	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"
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

// NewFluidControllerRateLimiter inherits the default controller rate limiter in workqueue.DefaultControllerRateLimiter()
// but with configurable parameters passed to the underlying rate limiters.
func NewFluidControllerRateLimiter(
	ItemExponentialFailureBaseDelay time.Duration,
	ItemExponentialFailureMaxDelay time.Duration,
	overallBucketQPS int,
	overallBucketBurst int) ratelimiter.RateLimiter {
	return workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(ItemExponentialFailureBaseDelay, ItemExponentialFailureMaxDelay),
		// 10 qps, 100 bucket size.  This is only for retry speed and its only the overall factor (not per item)
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(overallBucketQPS), overallBucketBurst)},
	)
}

// GetConfigOrDieWithQPSAndBurst sets client-side QPS and burst in kube config.
func GetConfigOrDieWithQPSAndBurst(qps float32, burst int) *rest.Config {
	cfg := ctrl.GetConfigOrDie()
	if qps > 0 {
		cfg.QPS = qps
	}

	if burst > 0 {
		cfg.Burst = burst
	}

	return cfg
}
