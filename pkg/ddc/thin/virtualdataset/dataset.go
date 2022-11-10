package virtualdataset

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

func getDatasetRefName(name, namespace string) string {
	// split by '#', can not use '-' because namespace or name can contain '-'
	return fmt.Sprintf("%s#%s", namespace, name)
}

func getMountedDatasetNamespacedName(virtualDataset *datav1alpha1.Dataset) []types.NamespacedName {
	// virtual dataset can only mount dataset
	var physicalNameSpacedName []types.NamespacedName
	for _, mount := range virtualDataset.Spec.Mounts {
		if common.IsFluidRefSchema(mount.MountPoint) {
			datasetPath := strings.TrimPrefix(mount.MountPoint, string(common.RefSchema))
			namespaceAndName := strings.Split(datasetPath, "/")
			physicalNameSpacedName = append(physicalNameSpacedName, types.NamespacedName{
				Namespace: namespaceAndName[0],
				Name:      namespaceAndName[1],
			})
		}
	}
	return physicalNameSpacedName
}
