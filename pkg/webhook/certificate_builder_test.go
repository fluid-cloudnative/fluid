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

package webhook

import (
	"context"
	"os"
	"reflect"
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

func TestBuildOrSyncCABundle(t *testing.T) {

	testCases := map[string]struct {
		ns          string
		svc         string
		certPath    string
		clientIsNil bool
	}{
		"test build and sync ca case 1": {
			ns:          common.NamespaceFluidSystem,
			svc:         "fluid-pod-admission-webhook",
			certPath:    "fluid_certs1",
			clientIsNil: false,
		},
		"test build and sync ca case 2": {
			ns:          common.NamespaceFluidSystem,
			svc:         "fluid-deployment-admission-webhook",
			certPath:    "fluid_certs2",
			clientIsNil: false,
		},
		"test build and sync ca case 3": {
			ns:          "kube-system",
			svc:         "fluid-pod-admission-webhook",
			certPath:    "fluid_certs3",
			clientIsNil: true,
		},
		"test build and sync ca case 4": {
			ns:          "default",
			svc:         "fluid-statefulSet-admission-webhook",
			certPath:    "fluid_certs4",
			clientIsNil: true,
		},
		"test build and sync ca case 5": {
			ns:          "",
			svc:         "fluid-pod-admission-webhook",
			certPath:    "fluid_certs3",
			clientIsNil: false,
		},
	}
	for _, item := range testCases {
		if item.ns != "" {
			t.Setenv(common.MyPodNamespace, item.ns)
		}

		certDir, err := os.MkdirTemp("/tmp", item.certPath)
		if err != nil {
			t.Errorf("MkdirTemp failed due to %v", err)
		}

		client := fake.NewFakeClientWithScheme(testScheme)
		cb := NewCertificateBuilder(client, log)
		if item.clientIsNil {
			cb.Client = nil
		}
		caCert, err := cb.BuildOrSyncCABundle(item.svc, certDir)
		if err != nil {
			if item.clientIsNil || item.ns == "" {
				continue
			}
			t.Errorf("fail to build or sync ca, err:%v", err)
		}

		// check if the cert files are generated correctly
		_, err = os.Stat(certDir + "/ca-key.pem")
		switch {
		case os.IsNotExist(err):
			t.Errorf("ca-key.pem not exist in certpath %v", certDir)
		case err != nil:
			t.Errorf("fail to check if ca-key.pem exist because of err %v", err)
		}
		_, err = os.Stat(certDir + "/ca-cert.pem")
		switch {
		case os.IsNotExist(err):
			t.Errorf("ca-cert.pem not exist in certpath %v", certDir)
		case err != nil:
			t.Errorf("fail to check if ca-cert.pem exist because of err %v", err)
		}
		_, err = os.Stat(certDir + "/cert.pem")
		switch {
		case os.IsNotExist(err):
			t.Errorf("cert.pem not exist in certpath %v", certDir)
		case err != nil:
			t.Errorf("fail to check if cert.pem exist because of err %v", err)
		}
		_, err = os.Stat(certDir + "/key.pem")
		switch {
		case os.IsNotExist(err):
			t.Errorf("key.pem not exist in certpath %v", certDir)
		case err != nil:
			t.Errorf("fail to check if key.pem exist because of err %v", err)
		}
		_, err = os.Stat(certDir + "/tls.crt")
		switch {
		case os.IsNotExist(err):
			t.Errorf("tls.crt not exist in certpath %v", certDir)
		case err != nil:
			t.Errorf("fail to check if tls.crt exist because of err %v", err)
		}
		_, err = os.Stat(certDir + "/tls.key")
		switch {
		case os.IsNotExist(err):
			t.Errorf("tls.key not exist in certpath %v", certDir)
		case err != nil:
			t.Errorf("fail to check if tls.key exist because of err %v", err)
		}

		// check if the ca-cert.pem file is the same as the return value of the function
		content, err := os.ReadFile(certDir + "/ca-cert.pem")
		if err != nil {
			t.Errorf("fail to read ca-cert.pem because of err %v", err)
		}
		if len(content) != len(caCert) {
			t.Errorf("the content of ca-cert.pem is %v, but the function return %v", content, caCert)
		} else {
			for i := 0; i < len(content); i++ {
				if content[i] != caCert[i] {
					t.Errorf("the content of ca-cert.pem is %v, but the function return %v", content, caCert)
				}

			}

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
					t.Errorf("%s cannot paas because fail to mutate CABundle of MutatingWebhookConfiguration", index)
					continue
				}
			}

		}

		err = cb.PatchCABundle(item.webhookName, item.ca)
		if err != nil {
			t.Errorf("%s cannot paas because fail to patch MutatingWebhookConfiguration", index)
		}
		var mc2 admissionregistrationv1.MutatingWebhookConfiguration
		err = client.Get(context.TODO(), types.NamespacedName{Name: mockWebhookName}, &mc2)
		if err != nil {
			t.Errorf("%s cannot paas because fail to get MutatingWebhookConfiguration", index)
		}
		if !reflect.DeepEqual(mc, mc2) {
			t.Errorf("should not patch MutatingWebhookConfiguration if not change")
		}
	}

}
