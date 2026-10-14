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

package transformer

import (
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var log = ctrl.Log.WithName("utils.transformer")

// fluidScheme knows the fluid API types. It recovers the GroupVersionKind of an object whose TypeMeta is not
// fully populated, which a typed client is allowed to hand back: an ownerReference missing its kind or its
// apiVersion is rejected by the API server and cannot be resolved back to its owner by an owner-based watch.
var fluidScheme = runtime.NewScheme()

func init() {
	utilruntime.Must(datav1alpha1.AddToScheme(fluidScheme))
}

func GenerateOwnerReferenceFromObject(obj client.Object) *common.OwnerReference {
	// The kind and the apiVersion fall back on their own, because a partially populated TypeMeta produces a
	// reference which is just as malformed as an entirely empty one.
	gvk := obj.GetObjectKind().GroupVersionKind()
	if len(gvk.Kind) == 0 || len(gvk.Version) == 0 || len(gvk.Group) == 0 {
		resolved, err := apiutil.GVKForObject(obj, fluidScheme)
		if err != nil {
			log.Error(err, "failed to recover the GroupVersionKind of the owner from the scheme, the generated ownerReference stays incomplete",
				"namespace", obj.GetNamespace(), "name", obj.GetName(), "groupVersionKind", gvk.String())
		} else {
			if len(gvk.Kind) == 0 {
				gvk.Kind = resolved.Kind
			}
			if len(gvk.Version) == 0 {
				gvk.Group, gvk.Version = resolved.Group, resolved.Version
			} else if len(gvk.Group) == 0 {
				// A group-less apiVersion such as "v1alpha1" parses with a non-empty version, so only the group is missing.
				gvk.Group = resolved.Group
			}
		}
	}

	ref := &common.OwnerReference{
		APIVersion:         gvk.GroupVersion().String(),
		Kind:               gvk.Kind,
		UID:                string(obj.GetUID()),
		Enabled:            true,
		Name:               obj.GetName(),
		BlockOwnerDeletion: false,
		Controller:         true,
	}

	return ref

}

func FilterOwnerByKind(ownerReferences []metav1.OwnerReference, ownerKind string) []metav1.OwnerReference {
	ret := []metav1.OwnerReference{}

	for _, owner := range ownerReferences {
		if owner.Kind == ownerKind {
			ret = append(ret, owner)
		}
	}

	return ret
}
