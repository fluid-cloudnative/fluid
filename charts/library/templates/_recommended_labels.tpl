{{/*
Common labels. Follow the following guidance:
- https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/#labels
- https://helm.sh/docs/chart_best_practices/labels/#standard-labels
*/}}
{{- define "library.fluid.labels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
fluid.io/managed-by: fluid
{{- end }}
