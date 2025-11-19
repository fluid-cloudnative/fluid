package jindocache

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/util/retry"
)

func (e *JindoCacheEngine) transformResources(runtime *datav1alpha1.JindoRuntime, value *Jindo, userQuotas string) (err error) {
	err = e.transformMasterResources(runtime, value, userQuotas)
	if err != nil {
		return
	}
	err = e.transformWorkerResources(runtime, value, userQuotas)
	if err != nil {
		return
	}
	e.transformFuseResources(runtime, value)

	return
}

func (e *JindoCacheEngine) transformMasterResources(runtime *datav1alpha1.JindoRuntime, value *Jindo, userQuotas string) (err error) {
	value.Master.Resources = utils.TransformCoreV1ResourcesToInternalResources(runtime.Spec.Master.Resources)

	limitMemEnable := false
	if os.Getenv("USE_DEFAULT_MEM_LIMIT") == "true" {
		limitMemEnable = true
	}

	// set memory request for the larger
	if e.hasTieredStore(runtime) && e.getTieredStoreType(runtime) == 0 {
		quotaString := strings.TrimRight(userQuotas, "g")
		needUpdated := false
		if quotaString != "" {
			i, _ := strconv.Atoi(quotaString)
			if limitMemEnable && i > defaultMemLimit {
				// value.Master.Resources.Requests.Memory = defaultMetaSize
				defaultMetaSizeQuantity := resource.MustParse(defaultMetaSize)
				if runtime.Spec.Master.Resources.Requests == nil ||
					runtime.Spec.Master.Resources.Requests.Memory() == nil ||
					runtime.Spec.Master.Resources.Requests.Memory().IsZero() ||
					defaultMetaSizeQuantity.Cmp(*runtime.Spec.Master.Resources.Requests.Memory()) > 0 {
					needUpdated = true
				}

				if !runtime.Spec.Master.Resources.Limits.Memory().IsZero() &&
					defaultMetaSizeQuantity.Cmp(*runtime.Spec.Master.Resources.Limits.Memory()) > 0 {
					return fmt.Errorf("the memory meta store's size %v is greater than master limits memory %v",
						defaultMetaSizeQuantity, runtime.Spec.Master.Resources.Limits.Memory())
				}

				if needUpdated {
					if value.Master.Resources.Requests == nil {
						value.Master.Resources.Requests = make(common.ResourceList)
					}
					value.Master.Resources.Requests[corev1.ResourceMemory] = defaultMetaSize
					err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
						runtime, err := e.getRuntime()
						if err != nil {
							return err
						}
						runtimeToUpdate := runtime.DeepCopy()
						if len(runtimeToUpdate.Spec.Master.Resources.Requests) == 0 {
							runtimeToUpdate.Spec.Master.Resources.Requests = make(corev1.ResourceList)
						}
						runtimeToUpdate.Spec.Master.Resources.Requests[corev1.ResourceMemory] = defaultMetaSizeQuantity
						if !reflect.DeepEqual(runtimeToUpdate, runtime) {
							err = e.Client.Update(context.TODO(), runtimeToUpdate)
							if err != nil {
								if apierrors.IsConflict(err) {
									time.Sleep(3 * time.Second)
								}
								return err
							}
							time.Sleep(1 * time.Second)
						}

						return nil
					})

					if err != nil {
						return err
					}

				}
			}
		}
	}

	return nil
}

func (e *JindoCacheEngine) transformWorkerResources(runtime *datav1alpha1.JindoRuntime, value *Jindo, userQuotas string) (err error) {
	value.Worker.Resources = utils.TransformCoreV1ResourcesToInternalResources(runtime.Spec.Worker.Resources)

	// mem set request
	if e.hasTieredStore(runtime) && e.getTieredStoreType(runtime) == 0 {
		userQuotas = strings.ReplaceAll(userQuotas, "g", "Gi")
		needUpdated := false
		userQuotasQuantity := resource.MustParse(userQuotas)
		if runtime.Spec.Worker.Resources.Requests == nil ||
			runtime.Spec.Worker.Resources.Requests.Memory() == nil ||
			runtime.Spec.Worker.Resources.Requests.Memory().IsZero() ||
			userQuotasQuantity.Cmp(*runtime.Spec.Worker.Resources.Requests.Memory()) > 0 {
			needUpdated = true
		}

		if !runtime.Spec.Worker.Resources.Limits.Memory().IsZero() &&
			userQuotasQuantity.Cmp(*runtime.Spec.Worker.Resources.Limits.Memory()) > 0 {
			return fmt.Errorf("the memory tieredStore's size %v is greater than worker limits memory %v",
				userQuotasQuantity, runtime.Spec.Worker.Resources.Limits.Memory())
		}
		if needUpdated {
			if value.Worker.Resources.Requests == nil {
				value.Worker.Resources.Requests = make(common.ResourceList)
			}
			value.Worker.Resources.Requests[corev1.ResourceMemory] = userQuotas
			err = retry.RetryOnConflict(retry.DefaultBackoff, func() error {
				runtime, err := e.getRuntime()
				if err != nil {
					return err
				}
				runtimeToUpdate := runtime.DeepCopy()
				if len(runtimeToUpdate.Spec.Worker.Resources.Requests) == 0 {
					runtimeToUpdate.Spec.Worker.Resources.Requests = make(corev1.ResourceList)
				}
				runtimeToUpdate.Spec.Worker.Resources.Requests[corev1.ResourceMemory] = userQuotasQuantity
				if !reflect.DeepEqual(runtimeToUpdate, runtime) {
					err = e.Client.Update(context.TODO(), runtimeToUpdate)
					if err != nil {
						if apierrors.IsConflict(err) {
							time.Sleep(3 * time.Second)
						}
						return err
					}
					time.Sleep(1 * time.Second)
				}

				return nil
			})

			if err != nil {
				return err
			}
		}

	}

	return
}

func (e *JindoCacheEngine) transformFuseResources(runtime *datav1alpha1.JindoRuntime, value *Jindo) {
	value.Fuse.Resources = utils.TransformCoreV1ResourcesToInternalResources(runtime.Spec.Fuse.Resources)
}
