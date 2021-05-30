package kubeclient

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	testScheme *runtime.Scheme
)

func init() {
	testScheme = runtime.NewScheme()
	_ = v1.AddToScheme(testScheme)
}

func TestGetPVCNamesFromPod(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	pod := v1.Pod{}
	var pvcNamesWant []string
	for i := 1; i <= 30; i++ {
		switch rand.Intn(4) {
		case 0:
			pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
				Name: "volume" + strconv.Itoa(i),
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: "/tmp/data" + strconv.Itoa(i),
					},
				},
			})
		case 1:
			pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
				Name: "volume" + strconv.Itoa(i),
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: "pvc" + strconv.Itoa(i),
						ReadOnly:  true,
					},
				},
			})
			pvcNamesWant = append(pvcNamesWant, "pvc"+strconv.Itoa(i))
		case 2:
			pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
				Name: "volume" + strconv.Itoa(i),
				VolumeSource: v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				},
			})
		case 3:
			pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
				Name: "volume" + strconv.Itoa(i),
				VolumeSource: v1.VolumeSource{
					NFS: &v1.NFSVolumeSource{
						Server:   "172.0.0." + strconv.Itoa(i),
						Path:     "/data" + strconv.Itoa(i),
						ReadOnly: true,
					},
				},
			})
		}
	}
	pvcNames := GetPVCNamesFromPod(&pod)

	if !reflect.DeepEqual(pvcNames, pvcNamesWant) {
		t.Errorf("the result of GetPVCNamesFromPod is not right")
	}

}
