package eac

import (
	"errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/nas"
	"github.com/brahma-adshonor/gohook"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ctrl"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func newEACEngine(client client.Client, name string, namespace string) *EACEngine {
	runTimeInfo, _ := base.BuildRuntimeInfo(name, namespace, common.EACRuntimeType, datav1alpha1.TieredStore{})
	engine := &EACEngine{
		runtime:     &datav1alpha1.EACRuntime{},
		name:        name,
		namespace:   namespace,
		Client:      client,
		runtimeInfo: runTimeInfo,
		Log:         fake.NullLogger(),
	}
	engine.Helper = ctrl.BuildHelper(runTimeInfo, client, engine.Log)
	return engine
}

func TestShouldCheckUFS(t *testing.T) {
	dataSetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "check",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "eac://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
						EncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: AccessKeyIDName,
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "check",
										Key:  "id",
									},
								},
							},
							{
								Name: AccessKeySecretName,
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "check",
										Key:  "secret",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nocheck",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "eac://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
					},
				},
			},
		},
	}

	secretInputs := []v1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "check",
				Namespace: "fluid",
			},
			StringData: map[string]string{
				"id":     "123",
				"secret": "321",
			},
		},
	}

	objs := []runtime.Object{}
	for _, d := range dataSetInputs {
		objs = append(objs, d.DeepCopy())
	}
	for _, s := range secretInputs {
		objs = append(objs, s.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)

	e1 := newEACEngine(fakeClient, "check", "fluid")
	should, err := e1.ShouldCheckUFS()
	if !should || err != nil {
		t.Errorf("testcase Failed due to")
	}

	e2 := newEACEngine(fakeClient, "nocheck", "fluid")
	should, err = e2.ShouldCheckUFS()
	if should || err != nil {
		t.Errorf("testcase Failed")
	}
}

func TestPrepareUFS(t *testing.T) {
	mockSetDirQuotaCommon := func(e *EACEngine, mountInfo MountInfo) (response *nas.SetDirQuotaResponse, err error) {
		return nil, nil
	}
	mockSetDirQuotaError := func(e *EACEngine, mountInfo MountInfo) (response *nas.SetDirQuotaResponse, err error) {
		return nil, errors.New("other error")
	}
	wrappedUnhookSetDirQuota := func(e *EACEngine) {
		err := gohook.UnHookMethod(e, "setDirQuota")
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	dataSetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "check",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "eac://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
						EncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: AccessKeyIDName,
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "check",
										Key:  "id",
									},
								},
							},
							{
								Name: AccessKeySecretName,
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "check",
										Key:  "secret",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nocheck",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "eac://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
					},
				},
			},
		},
	}

	secretInputs := []v1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "check",
				Namespace: "fluid",
			},
			StringData: map[string]string{
				"id":     "123",
				"secret": "321",
			},
		},
	}

	objs := []runtime.Object{}
	for _, d := range dataSetInputs {
		objs = append(objs, d.DeepCopy())
	}
	for _, s := range secretInputs {
		objs = append(objs, s.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
	e := newEACEngine(fakeClient, "check", "fluid")

	err := gohook.HookMethod(e, "setDirQuota", mockSetDirQuotaCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = e.PrepareUFS()
	if err != nil {
		t.Errorf("testcase Failed")
	}
	wrappedUnhookSetDirQuota(e)

	err = gohook.HookMethod(e, "setDirQuota", mockSetDirQuotaError, nil)
	if err == nil {
		t.Fatal(err.Error())
	}

	err = e.PrepareUFS()
	if err == nil {
		t.Errorf("testcase Failed")
	}
	wrappedUnhookSetDirQuota(e)
}

func TestTotalFileNums(t *testing.T) {
	mockdescribeDirQuotaCommon := func(e *EACEngine, mountInfo MountInfo) (response *nas.DescribeDirQuotasResponse, err error) {
		return &nas.DescribeDirQuotasResponse{
			DirQuotaInfos: []nas.DirQuotaInfo{
				{UserQuotaInfos: []nas.UserQuotaInfo{
					{FileCountReal: 123},
				}},
			},
		}, nil
	}
	mockdescribeDirQuotaError := func(e *EACEngine, mountInfo MountInfo) (response *nas.DescribeDirQuotasResponse, err error) {
		return nil, errors.New("other error")
	}
	wrappedUnhookdescribeDirQuota := func(e *EACEngine) {
		err := gohook.UnHookMethod(e, "describeDirQuota")
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	dataSetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "check",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "eac://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
						EncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: AccessKeyIDName,
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "check",
										Key:  "id",
									},
								},
							},
							{
								Name: AccessKeySecretName,
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "check",
										Key:  "secret",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nocheck",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "eac://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
					},
				},
			},
		},
	}

	secretInputs := []v1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "check",
				Namespace: "fluid",
			},
			StringData: map[string]string{
				"id":     "123",
				"secret": "321",
			},
		},
	}

	objs := []runtime.Object{}
	for _, d := range dataSetInputs {
		objs = append(objs, d.DeepCopy())
	}
	for _, s := range secretInputs {
		objs = append(objs, s.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
	e := newEACEngine(fakeClient, "check", "fluid")

	err := gohook.HookMethod(e, "describeDirQuota", mockdescribeDirQuotaCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = e.TotalFileNums()
	if err != nil {
		t.Errorf("testcase Failed")
	}
	wrappedUnhookdescribeDirQuota(e)

	err = gohook.HookMethod(e, "describeDirQuota", mockdescribeDirQuotaError, nil)
	if err == nil {
		t.Fatal(err.Error())
	}

	_, err = e.TotalFileNums()
	if err == nil {
		t.Errorf("testcase Failed")
	}
	wrappedUnhookdescribeDirQuota(e)
}

func TestTotalStorageBytes(t *testing.T) {
	mockdescribeDirQuotaCommon := func(e *EACEngine, mountInfo MountInfo) (response *nas.DescribeDirQuotasResponse, err error) {
		return &nas.DescribeDirQuotasResponse{
			DirQuotaInfos: []nas.DirQuotaInfo{
				{UserQuotaInfos: []nas.UserQuotaInfo{
					{FileCountReal: 123},
				}},
			},
		}, nil
	}
	mockdescribeDirQuotaError := func(e *EACEngine, mountInfo MountInfo) (response *nas.DescribeDirQuotasResponse, err error) {
		return nil, errors.New("other error")
	}
	wrappedUnhookdescribeDirQuota := func(e *EACEngine) {
		err := gohook.UnHookMethod(e, "describeDirQuota")
		if err != nil {
			t.Fatal(err.Error())
		}
	}

	dataSetInputs := []*datav1alpha1.Dataset{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "check",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "eac://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
						EncryptOptions: []datav1alpha1.EncryptOption{
							{
								Name: AccessKeyIDName,
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "check",
										Key:  "id",
									},
								},
							},
							{
								Name: AccessKeySecretName,
								ValueFrom: datav1alpha1.EncryptOptionSource{
									SecretKeyRef: datav1alpha1.SecretKeySelector{
										Name: "check",
										Key:  "secret",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nocheck",
				Namespace: "fluid",
			},
			Spec: datav1alpha1.DatasetSpec{
				Mounts: []datav1alpha1.Mount{
					{
						MountPoint: "eac://volume-uuid.region.nas.aliyuncs.com:/test-fluid-3",
					},
				},
			},
		},
	}

	secretInputs := []v1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "check",
				Namespace: "fluid",
			},
			StringData: map[string]string{
				"id":     "123",
				"secret": "321",
			},
		},
	}

	objs := []runtime.Object{}
	for _, d := range dataSetInputs {
		objs = append(objs, d.DeepCopy())
	}
	for _, s := range secretInputs {
		objs = append(objs, s.DeepCopy())
	}

	fakeClient := fake.NewFakeClientWithScheme(testScheme, objs...)
	e := newEACEngine(fakeClient, "check", "fluid")

	err := gohook.HookMethod(e, "describeDirQuota", mockdescribeDirQuotaCommon, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	_, err = e.TotalStorageBytes()
	if err != nil {
		t.Errorf("testcase Failed")
	}
	wrappedUnhookdescribeDirQuota(e)

	err = gohook.HookMethod(e, "describeDirQuota", mockdescribeDirQuotaError, nil)
	if err == nil {
		t.Fatal(err.Error())
	}

	_, err = e.TotalStorageBytes()
	if err == nil {
		t.Errorf("testcase Failed")
	}
	wrappedUnhookdescribeDirQuota(e)
}
