/*
Copyright 2021 The Fluid Authors.

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

package controllers

import (
	"os"
	"time"

	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"
)

// NewFluidControllerClient creates client.Client according to the HELM_DRIVER env variable. It returns the default client when setting HELM_DRIVER=true,
// meaning users explicitly grant secret permissions to Fluid controllers. Otherwise, it returns a specific client.Client that utilizes informers as cache
// except for Secrets.
func NewFluidControllerClient(config *rest.Config, options client.Options) (client.Client, error) {
	if driver, exist := os.LookupEnv("HELM_DRIVER"); exist && driver == "secret" {
		return client.New(config, options)
	}

	return NewCacheClientBypassSecrets(config, options)
}

// NewCacheClientBypassSecrets creates a client querying kubernetes resources with cache(informers) except for Secrets.
// Secret is an exception for that we aim to trade performance for higher security(e.g. less rbac verbs on Secrets).
func NewCacheClientBypassSecrets(config *rest.Config, options client.Options) (client.Client, error) {
	options.Cache.DisableFor = append(options.Cache.DisableFor, &corev1.Secret{})

	return client.New(config, options)
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
