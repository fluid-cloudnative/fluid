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
