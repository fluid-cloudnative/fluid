package alluxio

import (
	"context"
	"fmt"
	"github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

var propertiesToCheck = []string{
	"alluxio.master.rpc.port",
	"alluxio.master.web.port",
	"alluxio.worker.rpc.port",
	"alluxio.worker.web.port",
	"alluxio.job.master.rpc.port",
	"alluxio.job.master.web.port",
	"alluxio.job.worker.rpc.port",
	"alluxio.job.worker.web.port",
	"alluxio.job.worker.data.port",
	"alluxio.proxy.web.port",
	"alluxio.master.embedded.journal.port",
	"alluxio.job.master.embedded.journal.port",
}

func GetReservedPorts(client client.Client) (ports []int, err error) {
	var datasets v1alpha1.DatasetList
	err = client.List(context.TODO(), &datasets)
	if err != nil {
		return nil, errors.Wrap(err, "can't list datasets when GetReservedPorts")
	}

	for _, dataset := range datasets.Items {
		if len(dataset.Status.Runtimes) != 0 {
			// Assume there is only one runtime with category "Accelerate"
			accelerateRuntime := dataset.Status.Runtimes[0]
			if accelerateRuntime.Type != "alluxio" {
				continue
			}
			configMapName := fmt.Sprintf("%s-%s-values", accelerateRuntime.Name, accelerateRuntime.Type)
			configMap, err := kubeclient.GetConfigmapByName(client, configMapName, accelerateRuntime.Namespace)
			if err != nil {
				return nil, errors.Wrap(err, "GetConfigMapByName when GetReservedPorts")
			}

			if configMap == nil {
				continue
			}

			reservedPorts, err := parsePortsFromConfigMap(configMap)
			if err != nil {
				return nil, errors.Wrap(err, "parsePortsFromConfigMap when GetReservedPorts")
			}
			ports = append(ports, reservedPorts...)
		}
	}
	return ports, nil
}

func parsePortsFromConfigMap(configMap *v1.ConfigMap) (ports []int, err error) {
	var value Alluxio
	if v, ok := configMap.Data["data"]; ok {
		if err := yaml.Unmarshal([]byte(v), &value); err != nil {
			return nil, err
		}
		for _, property := range propertiesToCheck {
			if portStr, ok := value.Properties[property]; ok {
				portInt, err := strconv.Atoi(portStr)
				if err != nil {
					return nil, err
				}
				ports = append(ports, portInt)
			}
		}
	}
	return ports, nil
}
