{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "fluid.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "fluid.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "fluid.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "fluid.labels" -}}
helm.sh/chart: {{ include "fluid.chart" . }}
{{ include "fluid.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "fluid.selectorLabels" -}}
app.kubernetes.io/name: {{ include "fluid.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "fluid.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "fluid.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/* assemble images for fluid control-plane images */}}
{{- define "fluid.controlplane.imageTransform" -}}
  {{- $imagePrefix := index . 0 -}}
  {{- $imageName := index . 1 -}}
  {{- $imageTag := index . 2 -}}
  {{- $root := index . 3 -}}

  {{- /* If any value is empty, return an error message */ -}}
  {{- if or (empty $imagePrefix) (empty $imageName) (empty $imageTag) -}}
    {{- fail "Error: imagePrefix, imageName, and imageTag must all be defined and non-empty." -}}
  {{- end -}}

  {{- printf "%s/%s:%s" $imagePrefix $imageName $imageTag -}}
{{- end -}}

{{/* assemble images for runtime images */}}
{{- define "fluid.runtime.imageTransform" -}}
  {{- $imagePrefix := index . 0 -}}
  {{- $imageName := index . 1 -}}
  {{- $imageTag := index . 2 -}}
  {{- $root := index . 3 -}}

  {{- /* If any value is empty, return an error message */ -}}
  {{- if or (empty $imagePrefix) (empty $imageName) (empty $imageTag) -}}
    {{- fail "Error: imagePrefix, imageName, and imageTag must all be defined and non-empty." -}}
  {{- end -}}

  {{- printf "%s/%s:%s" $imagePrefix $imageName $imageTag -}}
{{- end -}}

{{- define "fluid.controlplane.resources" -}}
{{- $ := index . 0 -}}
{{- $resources := index . 1 -}}
{{- if $resources -}}
{{ toYaml $resources }}
{{- end -}}
{{- end -}}

{{- define "fluid.controlplane.affinity" -}}
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: type
          operator: NotIn
          values:
          - virtual-kubelet
{{- end -}}

{{- define "fluid.namespace" -}}
{{- if .Values.namespace -}}
    {{ .Values.namespace }}
{{- else -}}
    {{ .Release.Namespace }}
{{- end -}}
{{- end -}}

{{- define "fluid.helmDriver" -}}
{{- if or (eq .Values.helmDriver "configmap") (eq .Values.helmDriver "secret") -}}
{{ .Values.helmDriver | quote }}
{{- else -}}
{{ fail "helmDriver must be either configmap or secret" }}
{{- end -}}
{{- end -}}

{{- define "fluid.helmDriver.rbacs" -}}
{{- if eq .Values.helmDriver "secret" }}
  - apiGroups:
    - ""
    resources:
    - secrets
    verbs:
    - get
    - list
    - watch
    - create
    - update
    - delete
{{- end -}}
{{- end -}}


{{/*
Check if feature gate DataflowAffinity is enabled in the featureGates.
*/}}
{{- define "fluid.dataflowAffinity.enabled" -}}
{{- $featureGates := splitList "," .Values.fluidapp.featureGates }}
{{- $found := false -}}
{{- range $idx, $featureGate := $featureGates }}
    {{- $featureGateKV := splitList "=" $featureGate }}
    {{- $key :=  trim (index $featureGateKV 0) }}
    {{- $value := trim (index $featureGateKV 1) }}
    {{- if and (eq $key "DataflowAffinity") (eq $value "true") -}}
        {{- $found = true -}}
    {{- end -}}
{{- end -}}
{{- $found -}}
{{- end -}}
