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

package thin

import (
	"fmt"
	"reflect"
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestThinEngine_transfromSecretsForPersistentVolumeClaimMounts(t *testing.T) {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "fluid",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName: "test-pv",
		},
		Status: corev1.PersistentVolumeClaimStatus{
			Phase: corev1.ClaimBound,
		},
	}

	pv := &corev1.PersistentVolume{
		ObjectMeta: v1.ObjectMeta{
			Name: "test-pv",
		},
		Spec: corev1.PersistentVolumeSpec{
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				CSI: &corev1.CSIPersistentVolumeSource{
					NodePublishSecretRef: &corev1.SecretReference{
						Name:      "my-secret",
						Namespace: "node-publish-secrets",
					},
					VolumeHandle: "test-pv",
					VolumeAttributes: map[string]string{
						"test-attr":  "true",
						"test-attr2": "foobar",
					},
				},
			},
		},
	}

	mySecret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      "my-secret",
			Namespace: "fluid",
		},
		StringData: map[string]string{
			"encryptedValue": "test",
		},
	}

	nodePublishSecret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      "my-secret",
			Namespace: "node-publish-secrets",
		},
		StringData: map[string]string{
			"encryptedValue": "test",
		},
	}

	client := fake.NewFakeClientWithScheme(testScheme, pvc, pv, nodePublishSecret)

	engine := ThinEngine{
		Client:      client,
		name:        "thin-test",
		namespace:   "fluid",
		runtimeType: common.ThinRuntime,
		Log:         fake.NullLogger(),
		runtime: &datav1alpha1.ThinRuntime{
			ObjectMeta: v1.ObjectMeta{
				Name:      "thin-test",
				Namespace: "fluid",
			},
		},
	}

	dataset := &datav1alpha1.Dataset{
		Spec: datav1alpha1.DatasetSpec{
			Mounts: []datav1alpha1.Mount{
				{
					MountPoint: "pvc://test-pvc",
				},
			},
		},
	}

	thinValue := &ThinValue{
		Fuse: Fuse{},
	}

	t.Run("testing transformSecretsForpersistentVolumeClaimMounts with CopyNodePublishSecretIfNotExists policy", func(t *testing.T) {
		if err := engine.transfromSecretsForPersistentVolumeClaimMounts(dataset, datav1alpha1.CopyNodePublishSecretAndMountIfNotExists, thinValue); err != nil {
			t.Fatalf("expect no error, but got error %v", err)
		}

		expectPublishSecretName := fmt.Sprintf("%s-%s-publish-secret", engine.name, engine.runtimeType)
		copiedSecret, err := kubeclient.GetSecret(engine.Client, expectPublishSecretName, engine.namespace)
		if err != nil {
			t.Fatalf("expect found copied secret \"%s\", but got error: %v", expectPublishSecretName, err)
		}

		if !reflect.DeepEqual(copiedSecret.StringData, nodePublishSecret.StringData) {
			t.Fatalf("expect copied secret \"%s\" has same content, but not equal", copiedSecret.Name)
		}

		if len(thinValue.Fuse.Volumes) != 1 {
			t.Fatalf("expect appended volumes to fuse, but got %v", thinValue.Fuse.Volumes)
		}

		if len(thinValue.Fuse.VolumeMounts) != 1 {
			t.Fatalf("expect appended volumeMounts to fuse, but got %v", thinValue.Fuse.VolumeMounts)
		}
	})

	thinValue = &ThinValue{
		Fuse: Fuse{},
	}
	engine.Client = fake.NewFakeClientWithScheme(testScheme, pvc, pv, nodePublishSecret, mySecret)

	t.Run("testing transformSecretsForpersistentVolumeClaimMounts with MountNodePublishSecretIfExists policy", func(t *testing.T) {
		if err := engine.transfromSecretsForPersistentVolumeClaimMounts(dataset, datav1alpha1.MountNodePublishSecretIfExists, thinValue); err != nil {
			t.Fatalf("expect no error, but got error %v", err)
		}

		if len(thinValue.Fuse.Volumes) != 1 {
			t.Fatalf("expect appended volumes to fuse, but got %v", thinValue.Fuse.Volumes)
		}

		if len(thinValue.Fuse.VolumeMounts) != 1 {
			t.Fatalf("expect appended volumeMounts to fuse, but got %v", thinValue.Fuse.VolumeMounts)
		}
	})

	thinValue = &ThinValue{
		Fuse: Fuse{},
	}

	engine.Client = fake.NewFakeClientWithScheme(testScheme, pvc, pv)

	t.Run("testing transformSecretsForpersistentVolumeClaimMounts with NotMountNodePublishSecret policy", func(t *testing.T) {
		if err := engine.transfromSecretsForPersistentVolumeClaimMounts(dataset, datav1alpha1.NotMountNodePublishSecret, thinValue); err != nil {
			t.Fatalf("expect no error, but got error %v", err)
		}

		if len(thinValue.Fuse.Volumes) != 0 {
			t.Fatalf("expect no modification to volumes of fuse, but got %v", thinValue.Fuse.Volumes)
		}

		if len(thinValue.Fuse.VolumeMounts) != 0 {
			t.Fatalf("expect no modification to volumeMounts of fuse, but got %v", thinValue.Fuse.VolumeMounts)
		}
	})

}
