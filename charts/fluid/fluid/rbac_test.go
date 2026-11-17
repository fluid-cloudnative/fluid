/*
Copyright 2026 The Fluid Authors.

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

package fluid

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	yaml "sigs.k8s.io/yaml"
)

type renderedRBACObject struct {
	Kind     string `json:"kind"`
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	RoleRef struct {
		Name string `json:"name"`
	} `json:"roleRef"`
}

var defaultClusterRoleNames = []string{
	"alluxioruntime-controller", "cacheruntime-controller", "fluid-csi-plugin",
	"dataset-controller", "efcruntime-controller", "fluidapp-controller",
	"jindoruntime-controller", "juicefsruntime-controller", "thinruntime-controller",
	"vineyardruntime-controller", "fluid-webhook", "fluid-crds-upgrade",
}

var defaultClusterRoleBindingNames = []string{
	"alluxioruntime-clusterrolebinding", "cacheruntime-clusterrolebinding", "fluid-csi-plugin",
	"dataset-clusterrolebinding", "efcruntime-clusterrolebinding", "fluidapp-clusterrolebinding",
	"jindoruntime-clusterrolebinding", "juicefsruntime-clusterrolebinding", "thinruntime-clusterrolebinding",
	"vineyardruntime-clusterrolebinding", "fluid-webhook-clusterrolebinding", "fluid-crds-upgrade",
}

func renderClusterRBAC(t *testing.T, prefix string) []renderedRBACObject {
	t.Helper()
	helm, err := exec.LookPath("helm")
	if err != nil {
		t.Skip("helm is not installed; CI runs this test in the Helm-enabled lint job")
	}
	args := []string{"template", "fluid-test", ".", "--namespace", "fluid-system"}
	if prefix != "" {
		args = append(args, "--set", "rbac.clusterScopedNamePrefix="+prefix)
	}
	output, err := exec.Command(helm, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("helm template failed: %v\n%s", err, output)
	}
	var objects []renderedRBACObject
	for _, document := range strings.Split(string(output), "---") {
		var obj renderedRBACObject
		if err := yaml.Unmarshal([]byte(document), &obj); err != nil {
			t.Fatalf("parse rendered manifest: %v\n%s", err, document)
		}
		if obj.Kind == "ClusterRole" || obj.Kind == "ClusterRoleBinding" {
			objects = append(objects, obj)
		}
	}
	return objects
}

func TestClusterScopedRBACNamesPreserveDefaultsAndSupportOptInPrefix(t *testing.T) {
	defaults := renderClusterRBAC(t, "")
	prefixed := renderClusterRBAC(t, "tenant-a")
	kustomizeNames := make(map[string]bool)
	files, err := filepath.Glob("../../../config/fluid/bases/role/*/rbac.yaml")
	if err != nil {
		t.Fatalf("glob Kustomize RBAC files: %v", err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read Kustomize RBAC file %s: %v", file, err)
		}
		for _, document := range strings.Split(string(data), "---") {
			var obj renderedRBACObject
			if err := yaml.Unmarshal([]byte(document), &obj); err != nil {
				t.Fatalf("parse Kustomize RBAC file %s: %v", file, err)
			}
			if obj.Kind == "ClusterRole" || obj.Kind == "ClusterRoleBinding" {
				kustomizeNames[obj.Kind+":"+obj.Metadata.Name] = true
			}
		}
	}
	if len(kustomizeNames) == 0 {
		t.Fatal("no Kustomize cluster RBAC names found")
	}
	helmNames := make(map[string]bool)
	for _, obj := range defaults {
		if obj.Kind == "ClusterRole" || obj.Kind == "ClusterRoleBinding" {
			helmNames[obj.Kind+":"+obj.Metadata.Name] = true
		}
	}
	for name := range kustomizeNames {
		if !helmNames[name] {
			t.Errorf("existing Kustomize resource %q has no matching default Helm name", name)
		}
	}
	roleNames := make([]string, 0)
	bindingNames := make([]string, 0)
	for _, obj := range defaults {
		switch obj.Kind {
		case "ClusterRole":
			roleNames = append(roleNames, obj.Metadata.Name)
		case "ClusterRoleBinding":
			bindingNames = append(bindingNames, obj.Metadata.Name)
		}
	}
	sort.Strings(roleNames)
	sort.Strings(bindingNames)
	wantRoles := append([]string(nil), defaultClusterRoleNames...)
	wantBindings := append([]string(nil), defaultClusterRoleBindingNames...)
	sort.Strings(wantRoles)
	sort.Strings(wantBindings)
	if !reflect.DeepEqual(roleNames, wantRoles) {
		t.Fatalf("default ClusterRole names changed: got %v, want %v", roleNames, wantRoles)
	}
	if !reflect.DeepEqual(bindingNames, wantBindings) {
		t.Fatalf("default ClusterRoleBinding names changed: got %v, want %v", bindingNames, wantBindings)
	}
	assertReferences := func(objects []renderedRBACObject, prefix string) {
		t.Helper()
		availableRoles := make(map[string]bool)
		for _, obj := range objects {
			if obj.Kind == "ClusterRole" {
				availableRoles[obj.Metadata.Name] = true
			}
		}
		for _, obj := range objects {
			if obj.Kind != "ClusterRoleBinding" {
				continue
			}
			if !availableRoles[obj.RoleRef.Name] {
				t.Errorf("binding %q references missing ClusterRole %q", obj.Metadata.Name, obj.RoleRef.Name)
			}
			if prefix != "" && (!strings.HasPrefix(obj.Metadata.Name, prefix+"-") || !strings.HasPrefix(obj.RoleRef.Name, prefix+"-")) {
				t.Errorf("prefixed binding %q has unprefixed roleRef %q", obj.Metadata.Name, obj.RoleRef.Name)
			}
		}
		if prefix != "" {
			for _, obj := range objects {
				if obj.Kind == "ClusterRole" && !strings.HasPrefix(obj.Metadata.Name, prefix+"-") {
					t.Errorf("ClusterRole %q is missing opt-in prefix %q", obj.Metadata.Name, prefix)
				}
			}
		}
	}
	assertReferences(defaults, "")
	assertReferences(prefixed, "tenant-a")
	if len(prefixed) != len(defaults) {
		t.Fatalf("prefix changed number of cluster RBAC objects: got %d, want %d", len(prefixed), len(defaults))
	}
}
