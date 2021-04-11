package volume

import (
	"github.com/fluid-cloudnative/fluid/pkg/common"
	"github.com/fluid-cloudnative/fluid/pkg/ddc/base"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HasDeprecatedPersistentVolumeName(client client.Client, runtime base.RuntimeInfoInterface, log logr.Logger) (deprecated bool, err error) {
	deprecated, err = kubeclient.IsPersistentVolumeExist(client, runtime.GetName(), common.ExpectedFluidAnnotations)
	if err != nil {
		log.Error(err, "Failed to check if deprecated PV exists", "expeceted PV name", runtime.GetName())
		return
	}

	if deprecated {
		log.Info("Found deprecated PV", "pv name", runtime.GetName())
	} else {
		log.Info("No deprecated PV found, create pv instead", "runtime", runtime.GetName())
	}

	return
}
