/*
  Copyright 2022 The Fluid Authors.

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

package watch

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"testing"
)

func TestJobShouldInQueue(t *testing.T) {
	type args struct {
		job *batchv1.Job
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "cronjob",
			args: args{
				job: &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"cronjob": "cron dataload",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "parallel job",
			args: args{
				job: &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"parallelism": "3",
						},
					},
				},
			},
			want: false,
		},
		{
			name: "operation job",
			args: args{
				job: &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := JobShouldInQueue(tt.args.job); got != tt.want {
				t.Errorf("JobShouldInQueue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_opJobEventHandler_onDeleteFunc1(t *testing.T) {
	type args struct {
		client.Object
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test",
			args: args{
				Object: &batchv1.Job{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &opJobEventHandler{}
			predicate := h.onDeleteFunc(nil)

			delJobEvent := event.DeleteEvent{
				Object: tt.args.Object,
			}

			if predicate(delJobEvent) != tt.want {
				t.Errorf("onDeleteFunc() = %v, want %v", predicate(delJobEvent), tt.want)
			}
		})
	}
}

func Test_opJobEventHandler_onCreateFunc(t *testing.T) {
	type args struct {
		client.Object
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "job",
			args: args{
				Object: &batchv1.Job{},
			},
			want: true,
		},
		{
			name: "not job",
			args: args{
				Object: &batchv1.CronJob{},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &opJobEventHandler{}
			predicate := h.onCreateFunc(nil)

			createJobEvent := event.CreateEvent{
				Object: tt.args.Object,
			}

			if predicate(createJobEvent) != tt.want {
				t.Errorf("onCreateFunc() = %v, want %v", predicate(createJobEvent), tt.want)
			}
		})
	}
}

func Test_opJobEventHandler_onUpdateFunc(t *testing.T) {
	type args struct {
		old client.Object
		new client.Object
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "same job",
			args: args{
				old: &batchv1.Job{},
				new: &batchv1.Job{},
			},
			want: false,
		},
		{
			name: "not same job",
			args: args{
				old: &batchv1.Job{},
				new: &batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						ResourceVersion: "r1",
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &opJobEventHandler{}
			predicate := h.onUpdateFunc(nil)

			updateEvent := event.UpdateEvent{
				ObjectOld: tt.args.old,
				ObjectNew: tt.args.new,
			}
			if predicate(updateEvent) != tt.want {
				t.Errorf("onUpdateFunc() = %v, want %v", predicate(updateEvent), tt.want)
			}
		})
	}
}
