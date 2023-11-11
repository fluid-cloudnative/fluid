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

package efc

//func TestShouldCheckUFS(t *testing.T) {
//	dataSetInputs := []*datav1alpha1.Dataset{
//		{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      "check",
//				Namespace: "fluid",
//			},
//			Spec: datav1alpha1.DatasetSpec{
//				Mounts: []datav1alpha1.Mount{
//					{
//						MountPoint: "nfs://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
//						EncryptOptions: []datav1alpha1.EncryptOption{
//							{
//								Name: AccessKeyIDName,
//								ValueFrom: datav1alpha1.EncryptOptionSource{
//									SecretKeyRef: datav1alpha1.SecretKeySelector{
//										Name: "check",
//										Key:  "id",
//									},
//								},
//							},
//							{
//								Name: AccessKeySecretName,
//								ValueFrom: datav1alpha1.EncryptOptionSource{
//									SecretKeyRef: datav1alpha1.SecretKeySelector{
//										Name: "check",
//										Key:  "secret",
//									},
//								},
//							},
//						},
//					},
//				},
//			},
//		},
//		{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      "nocheck",
//				Namespace: "fluid",
//			},
//			Spec: datav1alpha1.DatasetSpec{
//				Mounts: []datav1alpha1.Mount{
//					{
//						MountPoint: "nfs://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
//					},
//				},
//			},
//		},
//	}
//
//	secretInputs := []v1.Secret{
//		{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      "check",
//				Namespace: "fluid",
//			},
//			Data: map[string][]byte{
//				"id":     []byte("123"),
//				"secret": []byte("321"),
//			},
//		},
//	}
//
//	objs := []runtime.Object{}
//	for _, d := range dataSetInputs {
//		objs = append(objs, d.DeepCopy())
//	}
//	for _, s := range secretInputs {
//		objs = append(objs, s.DeepCopy())
//	}
//
//	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
//
//	e1 := newEFCEngine(fakeClient, "check", "fluid")
//	should, err := e1.ShouldCheckUFS()
//	if !should || err != nil {
//		t.Errorf("testcase Failed due to")
//	}
//
//	e2 := newEFCEngine(fakeClient, "nocheck", "fluid")
//	should, err = e2.ShouldCheckUFS()
//	if should || err != nil {
//		t.Errorf("testcase Failed")
//	}
//}
//
//func TestPrepareUFS(t *testing.T) {
//	mockSetDirQuotaCommon := func(mountInfo MountInfo) (response *nas.SetDirQuotaResponse, err error) {
//		return nil, nil
//	}
//	mockSetDirQuotaError := func(mountInfo MountInfo) (response *nas.SetDirQuotaResponse, err error) {
//		return nil, errors.New("other error")
//	}
//	wrappedUnhookSetDirQuota := func() {
//		err := gohook.UnHook(MountInfo.SetDirQuota)
//		if err != nil {
//			t.Fatal(err.Error())
//		}
//	}
//
//	dataSetInputs := []*datav1alpha1.Dataset{
//		{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      "check",
//				Namespace: "fluid",
//			},
//			Spec: datav1alpha1.DatasetSpec{
//				Mounts: []datav1alpha1.Mount{
//					{
//						MountPoint: "nfs://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
//						EncryptOptions: []datav1alpha1.EncryptOption{
//							{
//								Name: AccessKeyIDName,
//								ValueFrom: datav1alpha1.EncryptOptionSource{
//									SecretKeyRef: datav1alpha1.SecretKeySelector{
//										Name: "check",
//										Key:  "id",
//									},
//								},
//							},
//							{
//								Name: AccessKeySecretName,
//								ValueFrom: datav1alpha1.EncryptOptionSource{
//									SecretKeyRef: datav1alpha1.SecretKeySelector{
//										Name: "check",
//										Key:  "secret",
//									},
//								},
//							},
//						},
//					},
//				},
//			},
//		},
//		{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      "nocheck",
//				Namespace: "fluid",
//			},
//			Spec: datav1alpha1.DatasetSpec{
//				Mounts: []datav1alpha1.Mount{
//					{
//						MountPoint: "nfs://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
//					},
//				},
//			},
//		},
//	}
//
//	secretInputs := []v1.Secret{
//		{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      "check",
//				Namespace: "fluid",
//			},
//			Data: map[string][]byte{
//				"id":     []byte("123"),
//				"secret": []byte("321"),
//			},
//		},
//	}
//
//	objs := []runtime.Object{}
//	for _, d := range dataSetInputs {
//		objs = append(objs, d.DeepCopy())
//	}
//	for _, s := range secretInputs {
//		objs = append(objs, s.DeepCopy())
//	}
//
//	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
//	e := newEFCEngine(fakeClient, "check", "fluid")
//
//	err := gohook.Hook(MountInfo.SetDirQuota, mockSetDirQuotaCommon, nil)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//
//	err = e.PrepareUFS()
//	if err != nil {
//		t.Errorf("testcase Failed")
//	}
//	wrappedUnhookSetDirQuota()
//
//	err = gohook.Hook(MountInfo.SetDirQuota, mockSetDirQuotaError, nil)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//
//	err = e.PrepareUFS()
//	if err == nil {
//		t.Errorf("testcase Failed")
//	}
//	wrappedUnhookSetDirQuota()
//}
//
//func TestTotalFileNumsAndTotalStorageBytes(t *testing.T) {
//	mockDescribeDirQuotaCommon := func(mountInfo MountInfo) (response *nas.DescribeDirQuotasResponse, err error) {
//		return &nas.DescribeDirQuotasResponse{
//			DirQuotaInfos: []nas.DirQuotaInfo{
//				{UserQuotaInfos: []nas.UserQuotaInfo{
//					{FileCountReal: 123},
//				}},
//			},
//		}, nil
//	}
//	mockDescribeDirQuotaError := func(mountInfo MountInfo) (response *nas.DescribeDirQuotasResponse, err error) {
//		return nil, errors.New("other error")
//	}
//	wrappedUnhookDescribeDirQuota := func() {
//		err := gohook.UnHook(MountInfo.DescribeDirQuota)
//		if err != nil {
//			t.Fatal(err.Error())
//		}
//	}
//
//	dataSetInputs := []*datav1alpha1.Dataset{
//		{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      "check",
//				Namespace: "fluid",
//			},
//			Spec: datav1alpha1.DatasetSpec{
//				Mounts: []datav1alpha1.Mount{
//					{
//						MountPoint: "nfs://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
//						EncryptOptions: []datav1alpha1.EncryptOption{
//							{
//								Name: AccessKeyIDName,
//								ValueFrom: datav1alpha1.EncryptOptionSource{
//									SecretKeyRef: datav1alpha1.SecretKeySelector{
//										Name: "check",
//										Key:  "id",
//									},
//								},
//							},
//							{
//								Name: AccessKeySecretName,
//								ValueFrom: datav1alpha1.EncryptOptionSource{
//									SecretKeyRef: datav1alpha1.SecretKeySelector{
//										Name: "check",
//										Key:  "secret",
//									},
//								},
//							},
//						},
//					},
//				},
//			},
//		},
//		{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      "nocheck",
//				Namespace: "fluid",
//			},
//			Spec: datav1alpha1.DatasetSpec{
//				Mounts: []datav1alpha1.Mount{
//					{
//						MountPoint: "nfs://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
//					},
//				},
//			},
//		},
//	}
//
//	secretInputs := []v1.Secret{
//		{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      "check",
//				Namespace: "fluid",
//			},
//			Data: map[string][]byte{
//				"id":     []byte("123"),
//				"secret": []byte("321"),
//			},
//		},
//	}
//
//	objs := []runtime.Object{}
//	for _, d := range dataSetInputs {
//		objs = append(objs, d.DeepCopy())
//	}
//	for _, s := range secretInputs {
//		objs = append(objs, s.DeepCopy())
//	}
//
//	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
//	e := newEFCEngine(fakeClient, "check", "fluid")
//
//	err := gohook.Hook(MountInfo.DescribeDirQuota, mockDescribeDirQuotaCommon, nil)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//
//	_, err = e.TotalFileNums()
//	if err != nil {
//		t.Errorf("testcase Failed")
//	}
//	_, err = e.TotalStorageBytes()
//	if err != nil {
//		t.Errorf("testcase Failed")
//	}
//	wrappedUnhookDescribeDirQuota()
//
//	err = gohook.Hook(MountInfo.DescribeDirQuota, mockDescribeDirQuotaError, nil)
//	if err != nil {
//		t.Fatal(err.Error())
//	}
//
//	_, err = e.TotalFileNums()
//	if err == nil {
//		t.Errorf("testcase Failed")
//	}
//	_, err = e.TotalStorageBytes()
//	if err == nil {
//		t.Errorf("testcase Failed")
//	}
//	wrappedUnhookDescribeDirQuota()
//}
