/* ==================================================================
* Copyright (c) 2023,11.5.
* All rights reserved.
*
* Redistribution and use in source and binary forms, with or without
* modification, are permitted provided that the following conditions
* are met:
*
* 1. Redistributions of source code must retain the above copyright
* notice, this list of conditions and the following disclaimer.
* 2. Redistributions in binary form must reproduce the above copyright
* notice, this list of conditions and the following disclaimer in the
* documentation and/or other materials provided with the
* distribution.
* 3. All advertising materials mentioning features or use of this software
* must display the following acknowledgement:
* This product includes software developed by the xxx Group. and
* its contributors.
* 4. Neither the name of the Group nor the names of its contributors may
* be used to endorse or promote products derived from this software
* without specific prior written permission.
*
* THIS SOFTWARE IS PROVIDED BY xxx,GROUP AND CONTRIBUTORS
* ===================================================================
* Author: xiao shi jie.
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
