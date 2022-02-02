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

package utils

import (
	"reflect"
	"testing"
)

const podYaml string = `
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  containers:
    - name: nginx
      image: nginx
      volumeMounts:
        - mountPath: /data
          name: hbase-vol
  volumes:
    - name: hbase-vol
      persistentVolumeClaim:
        claimName: hbase
`

const tfjobYaml string = `
apiVersion: "kubeflow.org/v1"
kind: "TFJob"
metadata:
  name: "mnist"
  namespace: kubeflow
  annotations:
   fluid.io/serverless: true
spec:
  cleanPodPolicy: None 
  tfReplicaSpecs:
    Worker:
      replicas: 1 
      restartPolicy: Never
      template:
        spec:
          containers:
            - name: tensorflow
              image: gcr.io/kubeflow-ci/tf-mnist-with-summaries:1.0
              command:
                - "python"
                - "/var/tf_mnist/mnist_with_summaries.py"
                - "--log_dir=/train/logs"
                - "--learning_rate=0.01"
                - "--batch_size=150"
              volumeMounts:
                - mountPath: "/train"
                  name: "training"
          volumes:
            - name: "training"
              persistentVolumeClaim:
                claimName: "tfevent-volume"  
`

func TestFromRawToObject(t *testing.T) {
	type testcase struct {
		name    string
		content string
		expect  string
	}

	testcases := []testcase{
		{
			name:    "pod",
			content: podYaml,
			expect:  "v1.Pod",
		}, {
			name:    "tfjob",
			content: tfjobYaml,
			expect:  "unstructured.Unstructured",
		},
	}

	for _, testcase := range testcases {
		obj, err := FromRawToObject([]byte(testcase.content))
		if err != nil {
			t.Errorf("failed due to error %v", err)
		}

		ref := reflect.TypeOf(obj)
		if ref.Kind().String() == testcase.expect {
			t.Errorf("the testcase %s failed: the expected result is %s, not %s",
				testcase.name,
				testcase.expect,
				ref.Kind().String())
		}
		// gvk := obj.GetObjectKind().GroupVersionKind()
		// if gvk.Kind != testcase.expect {
		// 	t.Errorf("the expected result is %s, not %s", testcase.expect, gvk.Kind)
		// }
	}
}
