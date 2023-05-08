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

package initcopy

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CopyScriptName    = "fluid-copy-script.sh"
	CopyScriptPath    = "/" + CopyScriptName
	CopyVolName       = "fluid-copy-script"
	CopyConfigMapName = CopyVolName
)

var replacer = strings.NewReplacer("¬", "`")
var contentCopyScript = `#!/bin/bash

set -ex
set +e
source_mount_path=$1
target_mount_path=$2
files=$3

for file in $(echo $files | tr ',' ' '); do
    
    target_file=$target_mount_path$file
    target_dir=$(dirname "$target_file")

    # if folder not exist, create it
    if [ ! -d "$target_dir" ]; then
        mkdir -p "$target_dir"
    fi

    source_file=$source_mount_path$file

    count=0
    while ((count < 5))
    do
        # copy file form source to target
        cp -r "$source_file" "$target_file"

        if [ $? -eq 0 ]; then
            echo "copy $source_file to $target_file"
            break
        fi

        sleep 3
        count=¬expr $count + 1¬
    done
    
    # fail to copy
    if test $count -ge 5; then
        echo "fail to copy $source_file to $target_file"
        exit 1
    fi    
done

echo "copy all sucessfully!"
`

type ScriptGeneratorForApp struct {
	namespace string
}

func NewScriptGeneratorForApp(namespace string) *ScriptGeneratorForApp {
	return &ScriptGeneratorForApp{
		namespace: namespace,
	}
}

func (a *ScriptGeneratorForApp) BuildConfigmap(ownerReference metav1.OwnerReference) *corev1.ConfigMap {
	data := map[string]string{}
	content := contentCopyScript

	data[CopyScriptName] = replacer.Replace(content)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            a.getConfigmapName(),
			Namespace:       a.namespace,
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Data: data,
	}
}

func (a *ScriptGeneratorForApp) getConfigmapName() string {
	return CopyConfigMapName
}
