{{/*
Common component in Dataload-cronJob.Spec.Template.Spec. Follow the following guidance:
*/}}
{{- define "library.fluid.dataload.cronJobCommonTemplateSpec" -}}
{{- if .Values.dataloader.schedulerName }}
schedulerName: {{ .Values.dataloader.schedulerName }}
{{- end }}
{{- with .Values.dataloader.nodeSelector }}
nodeSelector:
{{- toYaml . | nindent 2 }}
{{- end }}
{{- with .Values.dataloader.affinity }}
affinity:
{{- toYaml . | nindent 2 }}
{{- end }}
{{- with .Values.dataloader.tolerations }}
tolerations:
{{- toYaml . | nindent 2 }}
{{- end }}
restartPolicy: Never
{{- with .Values.dataloader.imagePullSecrets }}
imagePullSecrets:
{{- toYaml . | nindent 2 }}
{{- end }}
{{- end }}
