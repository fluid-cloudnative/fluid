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

package alluxio

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fluid-cloudnative/fluid/pkg/utils"

	"github.com/fluid-cloudnative/fluid/pkg/utils/helm"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// shut down the Alluxio engine
func (e *AlluxioEngine) Shutdown() (err error) {
	if e.retryShutdown < e.gracefulShutdownLimits {
		err = e.cleanupCache()
		if err != nil {
			e.retryShutdown = e.retryShutdown + 1
			e.Log.Info("clean cache failed",
				// "engine", e,
				"retry times", e.retryShutdown)
			return
		}
	}

	if e.MetadataSyncDoneCh != nil {
		close(e.MetadataSyncDoneCh)
	}

	_, err = e.destroyWorkers(-1)
	if err != nil {
		return
	}

	err = e.destroyMaster()
	if err != nil {
		return
	}

	err = e.cleanAll()
	return err
}

// destroyMaster Destroies the master
func (e *AlluxioEngine) destroyMaster() (err error) {
	found := false
	found, err = helm.CheckRelease(e.name, e.namespace)
	if err != nil {
		return err
	}

	if found {
		err = helm.DeleteRelease(e.name, e.namespace)
		if err != nil {
			return
		}
	}
	return
}

// // Destroy the workers
// func (e *AlluxioEngine) destroyWorkers() error {
// 	return nil
// }

// cleanupCache cleans up the cache
func (e *AlluxioEngine) cleanupCache() (err error) {
	// TODO(cheyang): clean up the cache
	cacheStates, err := e.queryCacheStatus()
	if cacheStates.cached == "" {
		return
	}

	e.Log.Info("The cache before cleanup",
		"cached", cacheStates.cached,
		"cachedPercentage", cacheStates.cachedPercentage)

	cached, err := utils.FromHumanSize(cacheStates.cached)
	if err != nil {
		return err
	}

	if cached == 0 {
		e.Log.Info("No need to clean cache",
			"cached", cacheStates.cached,
			"cachedPercentage", cacheStates.cachedPercentage)
		return nil
	}

	err = e.invokeCleanCache("/")
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil
		} else if strings.Contains(err.Error(), "does not have a host assigned") {
			return nil
		}
		return err
	} else {
		e.Log.Info("Clean up the cache successfully")
	}

	time.Sleep(time.Duration(10 * time.Second))

	// ufs, cached, cachedPercentage, err = e.du()
	// if err != nil {
	// 	return
	// }

	// e.Log.Info("The cache after cleanup", "ufs", ufs,
	// 	"cached", cached,
	// 	"cachedPercentage", cachedPercentage)

	// if cached > 0 {
	// 	return fmt.Errorf("The remaining cached is not cleaned up, it still has %d", cached)
	// }

	return fmt.Errorf("the remaining cached is not cleaned up, check again")
}

// cleanAll cleans up the all
func (e *AlluxioEngine) cleanAll() (err error) {
	var (
		valueConfigmapName = e.name + "-" + e.runtimeType + "-values"
		configmapName      = e.name + "-config"
		namespace          = e.namespace
	)

	cms := []string{valueConfigmapName, configmapName}

	for _, cm := range cms {
		err = kubeclient.DeleteConfigMap(e.Client, cm, namespace)
		if err != nil {
			return
		}
	}

	return nil
}

// destroyWorkers will delete the workers by number of the workers, if workers is -1, it means all the workers are deleted
func (e *AlluxioEngine) destroyWorkers(numWorkersToDestroy int32) (numFail int32, err error) {
	var (
		nodeList *corev1.NodeList = &corev1.NodeList{}

		labelName          = e.getRuntimeLabelname()
		labelCommonName    = e.getCommonLabelname()
		labelMemoryName    = e.getStoragetLabelname(humanReadType, memoryStorageType)
		labelDiskName      = e.getStoragetLabelname(humanReadType, diskStorageType)
		labelTotalname     = e.getStoragetLabelname(humanReadType, totalStorageType)
		labelExclusiveName = utils.GetExclusiveKey()
	)

	labelNames := []string{labelName, labelTotalname, labelDiskName, labelMemoryName, labelCommonName}
	datasetLabels, err := labels.Parse(fmt.Sprintf("%s=true", labelCommonName))
	if err != nil {
		return
	}

	err = e.List(context.TODO(), nodeList, &client.ListOptions{
		LabelSelector: datasetLabels,
	})
	if err != nil {
		return
	}

	var nodes []corev1.Node
	if numWorkersToDestroy > 0 {
		// This is a scale in operation
		pvcMountNodes, err := kubeclient.GetPvcMountNodes(e.Client, e.name, e.namespace)
		if err != nil {
			return numWorkersToDestroy, err
		}

		worker2UsedCapacityMap, err := e.getWorkerUsedCapacity()
		if err != nil {
			return numWorkersToDestroy, err
		}

		// Scaling in nodes where there are pods using the related pvc can be dangerous.
		// Filter out such nodes
		for _, node := range nodeList.Items {
			if _, found := pvcMountNodes[node.Name]; !found {
				nodes = append(nodes, node)
			}
		}

		if len(nodes) >= 2 {
			// Sort candidate nodes by used capacity in ascending order
			sort.Slice(nodes, func(i, j int) bool {
				usageNodeA := getUsedCapacity(nodes[i], worker2UsedCapacityMap)
				usageNodeB := getUsedCapacity(nodes[j], worker2UsedCapacityMap)
				return usageNodeA < usageNodeB
			})
		}
	} else {
		// Destroy all workers
		nodes = nodeList.Items
	}

	// 1.select the nodes
	for _, node := range nodes {
		if numWorkersToDestroy == 0 {
			break
		}
		// nodes = append(nodes, &node)
		toUpdate := node.DeepCopy()
		if len(toUpdate.Labels) == 0 {
			continue
		}

		for _, label := range labelNames {
			delete(toUpdate.Labels, label)
		}

		exclusiveLabelValue := utils.GetExclusiveValue(e.namespace, e.name)
		if val, exist := toUpdate.Labels[labelExclusiveName]; exist && val == exclusiveLabelValue {
			delete(toUpdate.Labels, labelExclusiveName)
			labelNames = append(labelNames, labelExclusiveName)
		}

		if len(toUpdate.Labels) < len(node.Labels) {
			err := e.Client.Update(context.TODO(), toUpdate)
			if err != nil {
				return numWorkersToDestroy, err
			}
			e.Log.Info("Destory worker", "Dataset", e.name, "deleted worker node", node.Name, "removed labels", labelNames)
		}

		if numWorkersToDestroy != -1 {
			numWorkersToDestroy--
		}
	}

	return numWorkersToDestroy, nil
}

func (e *AlluxioEngine) getWorkerUsedCapacity() (map[string]int64, error) {
	// 2. run clean action
	capacityReport, err := e.reportCapacity()
	if err != nil {
		return nil, err
	}

	// An Example of capacityReport:
	//////////////////////////////////////
	// Capacity information for all workers:
	//    Total Capacity: 4096.00MB
	//        Tier: MEM  Size: 4096.00MB
	//    Used Capacity: 443.89MB
	//        Tier: MEM  Size: 443.89MB
	//    Used Percentage: 10%
	//    Free Percentage: 90%
	//
	// Worker Name      Last Heartbeat   Storage       MEM
	// 192.168.1.147    0                capacity      2048.00MB
	//                                   used          443.89MB (21%)
	// 192.168.1.146    0                capacity      2048.00MB
	//                                   used          0B (0%)
	/////////////////////////////////////
	lines := strings.Split(capacityReport, "\n")
	startIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "Worker Name") {
			startIdx = i + 1
			break
		}
	}

	// TODO(xuzhihao): error: no line starts with Worker Name

	worker2UsedCapacityMap := make(map[string]int64)
	lenLines := len(lines)
	for lineIdx := startIdx; lineIdx < lenLines; lineIdx += 2 {
		// e.g. ["192.168.1.147", "0", "capacity", "2048.00MB", "used", "443.89MB", "(21%)"]
		workerInfoFields := append(strings.Fields(lines[lineIdx]), strings.Fields(lines[lineIdx+1])...)
		workerName := workerInfoFields[0]
		usedCapacity, _ := utils.FromHumanSize(workerInfoFields[5])
		worker2UsedCapacityMap[workerName] = usedCapacity
	}

	return worker2UsedCapacityMap, nil
}

func getUsedCapacity(node corev1.Node, usedCapacityMap map[string]int64) int64 {
	var ip, hostname string
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			ip = addr.Address
		}
		if addr.Type == corev1.NodeInternalDNS {
			hostname = addr.Address
		}
	}

	if len(ip) != 0 {
		if usedCapacity, found := usedCapacityMap[ip]; found {
			return usedCapacity
		}
	}

	if len(hostname) != 0 {
		if usedCapacity, found := usedCapacityMap[hostname]; found {
			return usedCapacity
		}
	}
	// no info stored in Alluxio master. Scale in such node first.
	return 0
}
