{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "jindofs.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "jindofs.fullname" -}}
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
{{- define "jindofs.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Distribute credential key and values with secret volume mounting on Jindo's pods
*/}}
{{- define "jindofs.cred.secret.volumeMounts" -}}
- name: jindofs-secret-token
  mountPath: /token
  readOnly: true
{{- end -}}

{{/*
Distribute credential key and values with secret volumes
*/}}
{{- define "jindofs.cred.secret.volumes" -}}
{{- if .Values.UseStsToken }}
- name: jindofs-secret-token
  secret:
    secretName: {{ .Values.secret }}
{{- else }}
- name: jindofs-secret-token
  secret:
    secretName: {{ .Values.secret }}
    items:
    - key: {{ .Values.secretKey }}
      path: AccessKeyId
    - key: {{ .Values.secretValue }}
      path: AccessKeySecret
{{- end }}
{{- end -}}
