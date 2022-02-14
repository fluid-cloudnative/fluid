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

package webhook

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fluid-cloudnative/fluid/pkg/utils"

	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	testScheme *runtime.Scheme
	log        = ctrl.Log.WithName("test")
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
}

func TestNewCertificateBuilder(t *testing.T) {
	c := fake.NewFakeClient()
	cb := NewCertificateBuilder(c, log)
	if cb.log != log {
		t.Errorf("fail to new the CertificateBuilder because log is not coincident")
	}
	if cb.Client != c {
		t.Errorf("fail to new the CertificateBuilder because client is not coincident")
	}
}

func TestBuildAndSyncCABundle(t *testing.T) {
	var webhookName = "webhookName"
	var caBundles = [][]byte{
		{3, 5, 54, 34},
		{3, 8, 54, 4},
		{35, 5, 54, 4},
	}
	var testMutatingWebhookConfiguration = &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "webhook1",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: caBundles[0],
				},
			},
			{
				Name: "webhook2",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: caBundles[1],
				},
			},
			{
				Name: "webhook3",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: caBundles[2],
				},
			},
		},
	}
	// create dir
	certPath := "/tmp/fluid/certs"
	if err := os.MkdirAll(certPath, 0700); err != nil {
		t.Errorf("fail to create path, path:%s,err:%v", certPath, err)
	}
	if err := os.Setenv(common.MyPodNamespace, "default"); err != nil {
		t.Errorf("fail to set env of path, path:%s,err:%v", certPath, err)
	}

	testCases := map[string]struct {
		lengthCheck int
		ns          string
		svc         string
		clientIsNil bool
	}{
		"test build and sync ca case 1": {
			lengthCheck: 1000,
			ns:          "fluid-system",
			svc:         "fluid-pod-admission-webhook",
			clientIsNil: false,
		},
		"test build and sync ca case 2": {
			lengthCheck: 1000,
			ns:          "fluid-system",
			svc:         "fluid-deployment-admission-webhook",
			clientIsNil: false,
		},
		"test build and sync ca case 3": {
			lengthCheck: 1000,
			ns:          "fluid-system",
			svc:         "fluid-pod-admission-webhook",
			clientIsNil: true,
		},
		"test build and sync ca case 4": {
			lengthCheck: 1000,
			ns:          "fluid-system",
			svc:         "fluid-statefulSet-admission-webhook",
			clientIsNil: true,
		},
	}
	for index, item := range testCases {
		testScheme.AddKnownTypes(schema.GroupVersion{Group: "admissionregistration.k8s.io", Version: "v1"}, testMutatingWebhookConfiguration)
		client := fake.NewFakeClientWithScheme(testScheme, testMutatingWebhookConfiguration)
		cb := NewCertificateBuilder(client, log)
		if item.clientIsNil {
			cb.Client = nil
		}
		err, caCert := cb.BuildAndSyncCABundle(item.svc, webhookName, certPath)
		if err != nil {
			if item.clientIsNil {
				continue
			}
			t.Errorf("fail to build and sync ca, err:%v", err)
		}
		var mc admissionregistrationv1.MutatingWebhookConfiguration
		err = client.Get(context.TODO(), types.NamespacedName{Name: webhookName}, &mc)
		if err != nil {
			t.Errorf("%s cannot paas because fail to get MutatingWebhookConfiguration", index)
			continue
		}
		for i := range mc.Webhooks {
			if len(mc.Webhooks[i].ClientConfig.CABundle) < item.lengthCheck || len(mc.Webhooks[i].ClientConfig.CABundle) != len(caCert) {
				t.Errorf("%s generate certification failed, ns:%s,svc:%s,want greater than %v,got:%v",
					index,
					item.ns,
					item.svc,
					item.lengthCheck,
					len(mc.Webhooks[i].ClientConfig.CABundle),
				)
				continue
			}
			for j := range caCert {
				if mc.Webhooks[i].ClientConfig.CABundle[j] != caCert[j] {
					t.Errorf("%s generate certification failed, ns:%s,svc:%s, the return result is not consistent with the patch",
						index,
						item.ns,
						item.svc,
					)
				}
			}
		}

	}
}

func TestGenCA(t *testing.T) {

	testCases := map[string]struct {
		lengthCheck int
		ns          string
		svc         string
		certPath    string
		canMkDir    bool
	}{
		"test generate ca file case 1": {
			lengthCheck: 1000,
			ns:          "fluid-system",
			svc:         "fluid-pod-admission-webhook",
			certPath:    "fluid_certs",
			canMkDir:    true,
		},
		"test generate ca file case 2": {
			lengthCheck: 1000,
			ns:          "fluid-system",
			svc:         "fluid-pod-admission-webhook",
			certPath:    "TEST",
			canMkDir:    true,
		},
		"test generate ca file case 3": {
			lengthCheck: 1000,
			ns:          "fluid-system",
			svc:         "fluid-pod-admission-webhook",
			certPath:    "TEst",
			canMkDir:    false,
		},
	}

	c := fake.NewFakeClient()
	cb := NewCertificateBuilder(c, log)

	for index, item := range testCases {
		certDir, err := ioutil.TempDir("/tmp", item.certPath)
		if err != nil {
			t.Errorf("MkdirTemp failed due to %v", err)
		}
		certs, err := cb.genCA(item.ns, item.svc, certDir)
		if err != nil && !item.canMkDir {
			continue
		}
		gotLen := len(certs.CACert)
		if gotLen < item.lengthCheck {
			t.Errorf("%s generate certification failed, ns:%s,svc:%s,want greater than %v,got:%v",
				index,
				item.ns,
				item.svc,
				item.lengthCheck,
				gotLen,
			)
		}
	}

}

func TestPatchCABundle(t *testing.T) {
	var mockWebhookName = "mockWebhookName"
	testCases := map[string]struct {
		ca          []byte
		webhookName string
	}{
		"test case 1": {
			ca:          []byte{1, 2, 3},
			webhookName: "mockWebhookName",
		},
		"test case 2": {
			ca:          []byte{2, 3, 4},
			webhookName: "mockWebhookName",
		},
		"test case 3": {
			ca:          []byte{3, 4, 5},
			webhookName: "mockWebhookName",
		},
		"test case 4": {
			ca:          []byte{4, 5, 6},
			webhookName: "WebhookName",
		},
	}

	var testMutatingWebhookConfiguration = &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: mockWebhookName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name: "webhook1",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: []byte{3, 5, 54, 34},
				},
			},
			{
				Name: "webhook2",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: []byte{3, 8, 54, 4},
				},
			},
			{
				Name: "webhook3",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: []byte{35, 5, 54, 4},
				},
			},
		},
	}

	testScheme.AddKnownTypes(schema.GroupVersion{Group: "admissionregistration.k8s.io", Version: "v1"}, testMutatingWebhookConfiguration)
	client := fake.NewFakeClientWithScheme(testScheme, testMutatingWebhookConfiguration)

	for index, item := range testCases {
		cb := NewCertificateBuilder(client, log)
		err := cb.PatchCABundle(item.webhookName, item.ca)
		if err != nil {
			if utils.IgnoreNotFound(err) != nil {
				t.Errorf("%s cannot paas because fail to patch MutatingWebhookConfiguration", index)
			} else {
				continue
			}
		}
		var mc admissionregistrationv1.MutatingWebhookConfiguration
		err = client.Get(context.TODO(), types.NamespacedName{Name: mockWebhookName}, &mc)
		if err != nil {
			t.Errorf("%s cannot paas because fail to get MutatingWebhookConfiguration", index)
			continue
		}
		for i := range mc.Webhooks {
			if len(mc.Webhooks[i].ClientConfig.CABundle) != len(item.ca) {
				t.Errorf("%s cannot paas because fail to mutate CABundle ofmMutatingWebhookConfiguration", index)
				continue
			}
			for j := range item.ca {
				if mc.Webhooks[i].ClientConfig.CABundle[j] != item.ca[j] {
					t.Errorf("%s cannot paas because fail to mutate CABundle ofmMutatingWebhookConfiguration", index)
					continue
				}
			}

		}
	}

}
