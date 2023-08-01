package efc

import (
	"testing"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	cruntime "github.com/fluid-cloudnative/fluid/pkg/runtime"
	"github.com/fluid-cloudnative/fluid/pkg/utils/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestEFCEngine_genverateDataProcessValueFile(t *testing.T) {
	dataset := &datav1alpha1.Dataset{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-dataset",
			Namespace: "default",
		},
	}

	dataProcess := &datav1alpha1.DataProcess{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-dataprocess",
			Namespace: "default",
		},
		Spec: datav1alpha1.DataProcessSpec{
			Dataset: datav1alpha1.TargetDatasetWithMountPath{
				TargetDataset: datav1alpha1.TargetDataset{
					Name:      "demo-dataset",
					Namespace: "default",
				},
				MountPath: "/data",
			},
			Processor: datav1alpha1.Processor{
				Script: &datav1alpha1.ScriptProcessor{
					RestartPolicy: corev1.RestartPolicyNever,
					VersionSpec: datav1alpha1.VersionSpec{
						Image:    "test-image",
						ImageTag: "latest",
					},
				},
			},
		},
	}

	type args struct {
		engine *EFCEngine
		ctx    cruntime.ReconcileRequestContext
		object client.Object
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "TestNotOfTypeDataProcess",
			args: args{
				engine: &EFCEngine{},
				ctx:    cruntime.ReconcileRequestContext{},
				object: &datav1alpha1.Dataset{}, // not of type DataProcess
			},
			wantErr: true,
		},
		{
			name: "TestTargetDatasetNotFound",
			args: args{
				engine: &EFCEngine{
					Client: fake.NewFakeClientWithScheme(testScheme), // No dataset
				},
				ctx: cruntime.ReconcileRequestContext{},
				object: &datav1alpha1.DataProcess{
					Spec: datav1alpha1.DataProcessSpec{
						Dataset: datav1alpha1.TargetDatasetWithMountPath{
							TargetDataset: datav1alpha1.TargetDataset{
								Name:      "demo-dataset-notfound",
								Namespace: "default",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "TestGenerateDataProcessValueFile",
			args: args{
				engine: &EFCEngine{
					Client: fake.NewFakeClientWithScheme(testScheme, dataset),
				},
				ctx:    cruntime.ReconcileRequestContext{},
				object: dataProcess,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.args.engine.generateDataProcessValueFile(tt.args.ctx, tt.args.object)
			if (err != nil) != tt.wantErr {
				t.Errorf("EFCEngine.generateDataProcessValueFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
