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
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"text/template"
)

const fullnameInclude = `{{ include "fluid.fullname" . }}`

// renderFullname executes the fluid.fullname define from templates/_helpers.tpl
// for the given release name, so the tests check what the helper produces
// rather than only the template text.
func renderFullname(t *testing.T, releaseName string) string {
	t.Helper()
	helpers, err := os.ReadFile("templates/_helpers.tpl")
	if err != nil {
		t.Fatalf("failed to read helpers: %v", err)
	}
	src := strings.ReplaceAll(string(helpers), "\r\n", "\n")
	start := strings.Index(src, `{{- define "fluid.fullname" -}}`)
	if start < 0 {
		t.Fatalf("fluid.fullname helper not found in templates/_helpers.tpl")
	}
	block := src[start:]
	if next := strings.Index(block[1:], `{{- define "`); next >= 0 {
		block = block[:next+1]
	}
	funcs := template.FuncMap{
		"default": func(d, v interface{}) interface{} {
			if v == nil || v == "" {
				return d
			}
			return v
		},
		"contains":   func(substr, s string) bool { return strings.Contains(s, substr) },
		"trunc": func(n int, s string) string {
			if len(s) > n {
				return s[:n]
			}
			return s
		},
		"trimSuffix": func(suffix, s string) string { return strings.TrimSuffix(s, suffix) },
	}
	tmpl, err := template.New("helpers").Funcs(funcs).Parse(block)
	if err != nil {
		t.Fatalf("failed to parse fluid.fullname helper: %v", err)
	}
	data := map[string]interface{}{
		"Values":  map[string]interface{}{"fullnameOverride": nil, "nameOverride": nil},
		"Chart":   map[string]interface{}{"Name": "fluid"},
		"Release": map[string]interface{}{"Name": releaseName},
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "fluid.fullname", data); err != nil {
		t.Fatalf("failed to render fluid.fullname for release %q: %v", releaseName, err)
	}
	return buf.String()
}

// Cluster-scoped RBAC objects must be named from the Helm release, otherwise
// two Fluid releases in different namespaces cannot coexist on one cluster:
// Helm refuses to adopt a ClusterRole/ClusterRoleBinding owned by another release.
func TestAlluxioClusterScopedRBACNamesAreReleaseScoped(t *testing.T) {
	data, err := os.ReadFile("templates/role/alluxio/rbac.yaml")
	if err != nil {
		t.Fatalf("failed to read rbac template: %v", err)
	}
	clusterRole := fullnameInclude + `-alluxioruntime-controller`
	clusterRoleBinding := fullnameInclude + `-alluxioruntime-clusterrolebinding`

	found := 0
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	for _, doc := range strings.Split(content, "\n---\n") {
		switch {
		case strings.Contains(doc, "\nkind: ClusterRole\n"):
			found++
			if !strings.Contains(doc, "\n  name: "+clusterRole+"\n") {
				t.Errorf("ClusterRole name is not release-scoped: want metadata name %q", clusterRole)
			}
		case strings.Contains(doc, "\nkind: ClusterRoleBinding\n"):
			found++
			if !strings.Contains(doc, "\n  name: "+clusterRoleBinding+"\n") {
				t.Errorf("ClusterRoleBinding name is not release-scoped: want metadata name %q", clusterRoleBinding)
			}
			if !strings.Contains(doc, "\n  name: "+clusterRole+"\n") {
				t.Errorf("ClusterRoleBinding roleRef does not reference release-scoped ClusterRole %q", clusterRole)
			}
		}
	}
	if found != 2 {
		t.Fatalf("expected 2 cluster-scoped RBAC objects in alluxio rbac template, found %d", found)
	}

	// The helper must yield a distinct, release-derived prefix per release.
	a := renderFullname(t, "team-a")
	b := renderFullname(t, "team-b")
	if a != "team-a-fluid" || b != "team-b-fluid" {
		t.Fatalf("fluid.fullname does not follow the release name: got %q and %q", a, b)
	}
}

// Every ClusterRole and ClusterRoleBinding in the chart must take its name,
// and the roleRef name of its binding, from fluid.fullname.
func TestClusterScopedRBACNamesAreReleaseScoped(t *testing.T) {
	var files []string
	for _, pattern := range []string{"templates/role/*/*.yaml", "templates/upgrade/*.yaml"} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob %q: %v", pattern, err)
		}
		files = append(files, matches...)
	}
	if len(files) == 0 {
		t.Fatalf("no RBAC templates found")
	}

	docSplit := regexp.MustCompile(`(?m)^---\s*$`)
	nameLine := regexp.MustCompile(`(?m)^  name: (.+?)\s*$`)
	kindLine := regexp.MustCompile(`(?m)^kind: (\S+)\s*$`)
	checked := 0
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("failed to read %s: %v", f, err)
		}
		content := strings.ReplaceAll(string(data), "\r\n", "\n")
		for _, doc := range docSplit.Split(content, -1) {
			kind := kindLine.FindStringSubmatch(doc)
			if kind == nil || (kind[1] != "ClusterRole" && kind[1] != "ClusterRoleBinding") {
				continue
			}
			for _, m := range nameLine.FindAllStringSubmatch(doc, -1) {
				checked++
				if !strings.HasPrefix(m[1], fullnameInclude+"-") {
					t.Errorf("%s: %s name %q is not release-scoped", f, kind[1], m[1])
				}
			}
		}
	}
	if checked == 0 {
		t.Fatalf("no cluster-scoped RBAC names were checked")
	}
}
