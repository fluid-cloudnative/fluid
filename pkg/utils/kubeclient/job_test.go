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
package kubeclient

import (
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestGetJob(t *testing.T) {
	mockJobName := "fluid-test-job"
	mockJobNamespace := "default"
	initJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mockJobName,
			Namespace: mockJobNamespace,
		},
	}

	fakeClient := fake.NewFakeClient(initJob)

	testCases := map[string]struct {
		name      string
		namespace string
		wantName  string
		notFound  bool
	}{
		"test get DataLoad Job case 1": {
			name:      mockJobName,
			namespace: mockJobNamespace,
			wantName:  mockJobName,
			notFound:  false,
		},
		"test get DataLoad Job case 2": {
			name:      mockJobName + "not-exist",
			namespace: mockJobNamespace,
			wantName:  "",
			notFound:  true,
		},
	}

	for k, item := range testCases {
		gotJob, err := GetJob(fakeClient, item.name, item.namespace)
		if item.notFound {
			if err == nil && gotJob != nil {
				t.Errorf("%s check failure, want get err, but get nil", k)
			}
		} else {
			if gotJob.Name != item.wantName {
				t.Errorf("%s check failure, want DataLoad Job name:%s, got DataLoad Job name:%s", k, item.wantName, gotJob.Name)
			}
		}
	}
}
