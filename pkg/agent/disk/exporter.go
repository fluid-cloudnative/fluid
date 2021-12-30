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

package disk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/fluid-cloudnative/fluid/pkg/utils/kubeclient"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/types"
	"regexp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const defaultDeviceFilter = "^/dev/vd.*$"

var _ manager.Runnable = &exporter{}

type exporter struct {
	client client.Client
	nodeId string
}

func NewExporter(nodeId string, client client.Client) *exporter {
	return &exporter{client: client, nodeId: nodeId}
}

func (e *exporter) run() {
	node, err := kubeclient.GetNode(e.client, e.nodeId)
	if err != nil {
		panic(fmt.Sprintf("disk resource exporter cannot find node %s due to error %v", e.nodeId, err))
	}
	glog.Infof("Found node %s to be patched", node.Name)

	stats, err := getFilesystemStats()
	if err != nil {
		panic(err)
	}
	glog.Infof("Got %d filesystem stats", len(stats))

	deviceFilterPattern := regexp.MustCompile(defaultDeviceFilter)
	var patchItems []ExtendedResourcePatch
	for _, stat := range stats {
		if !deviceFilterPattern.MatchString(stat.device) {
			glog.V(1).Infof("Ignoring filesystem %s for device %s", stat.mountPoint, stat.device)
			continue
		}

		patchItems = append(patchItems, transformFilesystemToPatch(stat))
	}

	glog.Infof("Ready to patch: node %s, patches: %v", node.Name, patchItems)

	patchByteData, err := json.Marshal(patchItems)
	if err != nil {
		panic(err)
	}

	err = e.client.Status().Patch(context.TODO(), node, client.RawPatch(types.JSONPatchType, patchByteData))
	if err != nil {
		panic(err)
	}

	glog.Infof("Successfully patched node %s", node.Name)
}

func (e *exporter) Start(ctx context.Context) error {
	e.run()
	return nil
}
