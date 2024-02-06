{{/*
Common labels. Follow guidance at https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels
*/}}
{{- define "library.fluid.labels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: fluid
{{- end }}
