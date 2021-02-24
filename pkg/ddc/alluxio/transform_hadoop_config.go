package alluxio

import (
	"context"
	"fmt"
	datav1alpha1 "github.com/fluid-cloudnative/fluid/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

// transformHadoopConfig transforms the given value by checking existence of user-specific hadoop configurations
func (e *AlluxioEngine) transformHadoopConfig(runtime *datav1alpha1.AlluxioRuntime, value *Alluxio) (err error) {
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

	var confFiles []string
	for k := range hadoopConfigMap.Data {
		switch k {
		case HADOOP_CONF_HDFS_SITE_FILENAME:
			value.HadoopConfig.IncludeHdfsSite = true
			confFiles = append(confFiles, HADOOP_CONF_MOUNT_PATH+"/"+HADOOP_CONF_HDFS_SITE_FILENAME)
		case HADOOP_CONF_CORE_SITE_FILENAME:
			value.HadoopConfig.IncludeCoreSite = true
			confFiles = append(confFiles, HADOOP_CONF_MOUNT_PATH+"/"+HADOOP_CONF_CORE_SITE_FILENAME)
		}
	}

	// Neither hdfs-site.xml nor core-site.xml is found in the configMap
	if !value.HadoopConfig.IncludeCoreSite && !value.HadoopConfig.IncludeHdfsSite {
		err = fmt.Errorf("Neither \"%v\" nor \"%v\" is found in the specified configMap \"%v\" ", HADOOP_CONF_HDFS_SITE_FILENAME, HADOOP_CONF_CORE_SITE_FILENAME, runtime.Spec.HadoopConfig)
		return err
	}

	value.HadoopConfig.ConfigMap = runtime.Spec.HadoopConfig
	value.Properties["alluxio.underfs.hdfs.configuration"] = strings.Join(confFiles, ":")

	return nil
}
