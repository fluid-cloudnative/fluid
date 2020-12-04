package jindo

import (
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	"github.com/fluid-cloudnative/fluid/pkg/utils"
	"strings"
)

func (e *JindoEngine) transform(runtime *datav1alpha1.JindoRuntime) (value *Jindo, err error) {
	if runtime == nil {
		err = fmt.Errorf("the jindoRuntime is null")
		return
	}

	value = &Jindo{
		Master: Master{
			NodeSelector:     map[string]string{},
			MasterProperties: e.transformMaster(runtime, value),
		},
		Worker: Worker{
			NodeSelector:     map[string]string{},
			WorkerProperties: e.transformWorker(runtime, value),
		},
		Fuse: Fuse{
			Args:           nil,
			HostPath:       e.getMountPoint(),
			NodeSelector:   map[string]string{},
			FuseProperties: e.transformFuse(runtime, value),
		},
	}
	return value, nil
}

func (e *JindoEngine) transformMaster(runtime *datav1alpha1.JindoRuntime, value *Jindo) map[string]string {
	properties := map[string]string{}
	dataset, err := utils.GetDataset(e.Client, e.name, e.namespace)
	if err != nil {
		return properties
	}
	jfsNamespace := ""
	for _, mount := range dataset.Spec.Mounts {
		if !strings.HasPrefix(mount.MountPoint, "oss://") {
			continue
		}
		jfsNamespace = jfsNamespace + mount.Name + ","
		properties["jfs.namespaces."+mount.Name+".oss.uri"] = mount.MountPoint
		properties["jfs.namespaces."+mount.Name+".oss.access.key"] = mount.Options["fs.oss.accessKeyId"]
		properties["jfs.namespaces."+mount.Name+".oss.access.secret"] = mount.Options["fs.oss.accessKeySecret"]
		properties["jfs.namespaces."+mount.Name+".mode"] = "cache"
	}
	if strings.HasSuffix(jfsNamespace, ",") {
		jfsNamespace = strings.TrimRight(jfsNamespace, ",")
	}
	properties["jfs.namespaces"] = jfsNamespace
	// combine properties together
	if len(runtime.Spec.Master.Properties) > 0 {
		for k, v := range runtime.Spec.Master.Properties {
			properties[k] = v
		}
	}
	return properties
}

func (e *JindoEngine) transformWorker(runtime *datav1alpha1.JindoRuntime, value *Jindo) map[string]string {
	properties := map[string]string{}
	if len(runtime.Spec.Worker.Properties) > 0 {
		for k, v := range runtime.Spec.Worker.Properties {
			properties[k] = v
		}
	}
	return properties
}

func (e *JindoEngine) transformFuse(runtime *datav1alpha1.JindoRuntime, value *Jindo) map[string]int {
	properties := map[string]int{}
	if len(runtime.Spec.Worker.Properties) > 0 {
		for k, v := range runtime.Spec.Fuse.Properties {
			properties[k] = v
		}
	}
	return properties
}
