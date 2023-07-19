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

package jindocache

import (
	"context"
	"fmt"

	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// transformHadoopConfig transforms the given value by checking existence of user-specific hadoop configurations
func (e *JindoCacheEngine) transformHadoopConfig(runtime *datav1alpha1.JindoRuntime, value *Jindo) (err error) {
	if len(runtime.Spec.HadoopConfig) == 0 {
		return nil
	}

	key := types.NamespacedName{
		Namespace: runtime.Namespace,
		Name:      runtime.Spec.HadoopConfig,
	}

	hadoopConfigMap := &v1.ConfigMap{}

	if err = e.Client.Get(context.TODO(), key, hadoopConfigMap); err != nil {
		if apierrs.IsNotFound(err) {
			err = fmt.Errorf("specified hadoopConfig \"%v\" is not found", runtime.Spec.HadoopConfig)
		}
		return err
	}

	for k := range hadoopConfigMap.Data {
		switch k {
		case HADOOP_CONF_HDFS_SITE_FILENAME:
			value.HadoopConfig.IncludeHdfsSite = true
		case HADOOP_CONF_CORE_SITE_FILENAME:
			value.HadoopConfig.IncludeCoreSite = true
		}
	}

	// Neither hdfs-site.xml nor core-site.xml is found in the configMap
	if !value.HadoopConfig.IncludeCoreSite && !value.HadoopConfig.IncludeHdfsSite {
		err = fmt.Errorf("neither \"%v\" nor \"%v\" is found in the specified configMap \"%v\" ", HADOOP_CONF_HDFS_SITE_FILENAME, HADOOP_CONF_CORE_SITE_FILENAME, runtime.Spec.HadoopConfig)
		return err
	}

	value.HadoopConfig.ConfigMap = runtime.Spec.HadoopConfig

	return nil
}
