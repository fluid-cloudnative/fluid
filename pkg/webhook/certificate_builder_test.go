/*

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
	"os"
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGenCA(t *testing.T) {

	testCases := map[string]struct {
		lengthCheck int
		ns          string
		svc         string
	}{
		"test generate ca file case 1": {
			lengthCheck: 1000,
			ns:          "fluid-system",
			svc:         "fluid-pod-admission-webhook",
		},
	}

	certExeFile := "../../tools/certificate.sh"
	certPath := "/tmp/fluid/certs"

	// create dir
	if err := os.MkdirAll(certPath, 0700); err != nil {
		t.Errorf("fail to create path, path:%s,err:%v", certPath, err)
	}

	c := fake.NewFakeClient()
	cb := NewCertificateBuilder(c, ctrl.Log.WithName("test"))

	for index, item := range testCases {
		ca, _ := cb.genCA(item.ns, item.svc, certExeFile, certPath)
		gotLen := len(ca)
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

	// clean certificate files
	if err := os.RemoveAll(certPath); err != nil {
		t.Errorf("fail to recycle file, path:%s,err:%v", certPath, err)
	}
}
