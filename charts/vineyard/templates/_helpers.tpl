{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "vineyard.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "vineyard.fullname" -}}
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
{{- define "vineyard.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/*
Define the resources for vineyard master.
*/}}
{{- define "vineyard.master.resources" -}}
resources:
  limits:
    {{- if .Values.master.resources.limits }}
      {{- if .Values.master.resources.limits.cpu  }}
    cpu: {{ .Values.master.resources.limits.cpu }}
      {{- end }}
      {{- if .Values.master.resources.limits.memory  }}
    memory: {{ .Values.master.resources.limits.memory }}
      {{- end }}
    {{- end }}
  requests:
    {{- if .Values.master.resources.requests }}
      {{- if .Values.master.resources.requests.cpu  }}
    cpu: {{ .Values.master.resources.requests.cpu }}
      {{- end }}
      {{- if .Values.master.resources.requests.memory  }}
    memory: {{ .Values.master.resources.requests.memory }}
      {{- end }}
    {{- end }}
{{- end -}}


{{/*
Increase the memory by 500Mi on top of the shared memory to meet the requirements for running Vineyard.
*/}}
{{- define "vineyard.addVineyardMemory" -}}
{{- $input := . -}}
{{- $numberPart := regexFind "^[0-9]+(\\.[0-9]+)?" $input -}}
{{- $unit := regexFind "[A-Za-z]+$" $input -}}
{{- $integerPart := regexFind "^[0-9]+" $numberPart -}}

{{- $decimalPart := "" -}}
{{- if regexMatch "^[0-9]+\\.[0-9]+" $numberPart -}}
  {{- $decimalPart = regexFind "\\.[0-9]+" $numberPart -}}
  {{- $decimalPart = trimPrefix "." $decimalPart -}}
{{- end -}}

{{- $decimalLength := len $decimalPart -}}
{{- if lt $decimalLength 3 -}}
  {{- if eq $decimalLength 2 -}}
    {{- $decimalPart = printf "%s0" $decimalPart -}}
  {{- else if eq $decimalLength 1 -}}
    {{- $decimalPart = printf "%s00" $decimalPart -}}
  {{- else -}}
    {{- $decimalPart = "000" -}}
  {{- end -}}
{{- end -}}

{{- $intValue := atoi $integerPart -}}
{{- $decimalAsInt := atoi $decimalPart -}}
{{- $valueAsMi := 0 -}}
{{- $adjustedValue := 0 -}}
{{- if or (eq $unit "Gi") (eq $unit "G") -}}
  {{- $valueAsMi = add (mul $intValue 1024) $decimalPart -}}
  {{- $adjustedValue = add $valueAsMi 500 -}}
{{- else if or (eq $unit "Mi") (eq $unit "M") -}}
  {{- $adjustedValue = add (atoi $numberPart) 500 -}}
{{- end -}}
{{ printf "%dMi" $adjustedValue -}}
{{- end -}}


{{/*
Define the resources for vineyard worker.
*/}}
{{- define "vineyard.worker.resources" -}}
{{- $size := include "vineyard.worker.size" . -}}
{{- $newSize := include "vineyard.addVineyardMemory" $size -}}
resources:
  limits:
    {{- if .Values.worker.resources.limits }}
      {{- if .Values.worker.resources.limits.cpu }}
    cpu: {{ .Values.worker.resources.limits.cpu }}
      {{- end }}
      {{- $memoryLimit := default $newSize .Values.worker.resources.limits.memory }}
      {{- if or (not .Values.worker.resources.limits.memory) (lt .Values.worker.resources.limits.memory $size) }}
    memory: {{ $memoryLimit }}
      {{- end }}
    {{- else }}
    memory: {{ $newSize }}
    {{- end }}
  requests:
    {{- if .Values.worker.resources.requests }}
      {{- if .Values.worker.resources.requests.cpu }}
    cpu: {{ .Values.worker.resources.requests.cpu }}
      {{- end }}
      {{- $memoryRequest := default $newSize .Values.worker.resources.requests.memory }}
      {{- if or (not .Values.worker.resources.requests.memory) (lt .Values.worker.resources.requests.memory $size) }}
    memory: {{ $memoryRequest }}
      {{- end }}
    {{- else }}
    memory: {{ $newSize }}
    {{- end }}
{{- end -}}


{{/*
Define the resources for vineyard fuse.
*/}}
{{- define "vineyard.fuse.resources" -}}
resources:
  limits:
    {{- if .Values.fuse.resources.limits }}
      {{- if .Values.fuse.resources.limits.cpu  }}
    cpu: {{ .Values.fuse.resources.limits.cpu }}
      {{- end }}
      {{- if .Values.fuse.resources.limits.memory  }}
    memory: {{ .Values.fuse.resources.limits.memory }}
      {{- end }}
    {{- end }}
  requests:
    {{- if .Values.fuse.resources.requests }}
      {{- if .Values.fuse.resources.requests.cpu  }}
    cpu: {{ .Values.fuse.resources.requests.cpu }}
      {{- end }}
      {{- if .Values.fuse.resources.requests.memory  }}
    memory: {{ .Values.fuse.resources.requests.memory }}
      {{- end }}
    {{- end }}
{{- end -}}


{{/*
Generate etcd endpoints for peer or client
*/}}
{{- define "vineyard.etcdEndpoints" -}}
  {{- $replicas := int .Values.master.replicas }}
  {{- $fullName := include "vineyard.fullname" . }}
  {{- $etcdFullname := (printf "%s-%s" $fullName "master") }}
  {{- $releaseNamespace := .Release.Namespace }}
  {{- $etcdServiceName := (printf "%s-%s" $fullName "master-svc") }}
  {{- $etcdEndpoint := list }}
  {{- $portType := .portType }}
  {{- $port := 0 }}
  {{- if eq $portType "peer" }}
    {{- $port = int .Values.master.ports.peer }}
    {{- range $e, $i := until $replicas }}
      {{- $etcdEndpoint = append $etcdEndpoint (printf "%s-%d=http://%s-%d.%s.%s:%d" $etcdFullname $i $etcdFullname $i $etcdServiceName $releaseNamespace $port) }}
    {{- end }}
  {{- else if eq $portType "client" }}
    {{- $port = int .Values.master.ports.client }}
    {{- range $e, $i := until $replicas }}
      {{- $etcdEndpoint = append $etcdEndpoint (printf "http://%s-%d.%s.%s:%d" $etcdFullname $i $etcdServiceName $releaseNamespace $port) }}
    {{- end }}
  {{- end }}
  {{- join "," $etcdEndpoint }}
{{- end }}

{{/*
Get the master endpoint.
*/}}
{{- define "vineyard.master.endpoint" -}}
    {{- if .Values.worker.externalEndpoint -}}
        {{- printf "http://%s" .Values.worker.externalEndpoint.uri }}
    {{- else -}}
      {{ include "vineyard.etcdEndpoints" (dict "Values" .Values "Release" .Release "portType" "client") }}
    {{- end -}}
{{- end -}}


{{/*
Get the shared memory size of the worker.
*/}}
{{- define "vineyard.worker.size" -}}
{{- $quota := "4Gi" -}}
{{- range .Values.tieredstore.levels }}
  {{- if (eq .mediumtype "MEM") }}
    {{- $quota = .quota -}}
    {{- break -}}
  {{- end -}}
{{- end -}}
{{- $quota -}}
{{- end -}}


{{/*
Check the tiered store to find whether the spill mechanism is enabled.
*/}}
{{- define "vineyard.checkTieredStore" -}}
{{- $levels := index .Values "tieredstore" "levels" -}}
{{- $memLevel := dict -}}
{{- $ssdLevel := dict -}}
{{- range $index, $level := $levels -}}
  {{- if eq $level.mediumtype "MEM" -}}
    {{- $memLevel = $level -}}
  {{- end -}}
  {{- if eq $level.mediumtype "SSD" -}}
    {{- $ssdLevel = $level -}}
  {{- end -}}
{{- end -}}
{{- $memLevelOk := and $memLevel (hasKey $memLevel "high") (hasKey $memLevel "low") -}}
{{- $ssdLevelOk := and $ssdLevel (hasKey $ssdLevel "path") -}}
{{- $valid := and $memLevelOk $ssdLevelOk -}}
{{- $valid -}}
{{- end -}}

{{/*
Get spill upper rate from tieredstore.
*/}}
{{- define "vineyard.spill.upperRate" -}}
{{- $high := "" -}}
{{- range .Values.tieredstore.levels -}}
  {{- if eq .mediumtype "MEM" -}}
    {{- $high = .high -}}
  {{- end -}}
{{- end -}}
{{- $high -}}
{{- end -}}


{{/*
Get spill lower rate from tieredstore.
*/}}
{{- define "vineyard.spill.lowerRate" -}}
{{- $low := "" -}}
{{- range .Values.tieredstore.levels -}}
  {{- if eq .mediumtype "MEM" -}}
    {{- $low = .low -}}
  {{- end -}}
{{- end -}}
{{- $low -}}
{{- end -}}


{{/*
Get spill path from tieredstore.
*/}}
{{- define "vineyard.spill.path" -}}
{{- $path := "" -}}
{{- range .Values.tieredstore.levels -}}
  {{- if eq (int .level) 1 -}}
    {{- $path = .path -}}
  {{- end -}}
{{- end -}}
{{- $path -}}
{{- end -}}
  

{{/*
Get spill volume Mounts from tieredstore.
*/}}
{{- define "vineyard.worker.tieredstoreVolumeMounts" -}}
  {{- $found := false -}}
  {{- range .Values.tieredstore.levels }}
    {{- if and (not $found) (eq .mediumtype "SSD") }}
          - mountPath: {{ .path | quote }}
            name: {{ .mediumtype | lower | quote }}
      {{- $found = true -}}
    {{- end }}
  {{- end }}
{{- end -}}


{{/*
Get spill volume from tieredstore.
*/}}
{{- define "vineyard.worker.tieredstoreVolume" -}}
  {{- if .Values.tieredstore.levels }}
    {{- range .Values.tieredstore.levels }}
      {{- if eq .mediumtype "SSD" }}
      {{- $mediumName := .mediumtype | lower }}
      {{- if eq .type "hostPath"}}
      - hostPath:
          path: {{ .path }}
          type: DirectoryOrCreate
        name: {{ $mediumName }}
      {{- else if eq .type "persistentVolumeClaim" }}
      - name: {{ $mediumName }}
        persistentVolumeClaim:
          claimName: {{ .name }}
      {{- else }}
      - name: {{ $mediumName }}
        emptyDir:
          medium: {{ eq .type "MEM" | ternary "Memory" "" }}
          {{- if .quota }}
          {{- /* quota should be transformed to match resource.Quantity. e.g. 20GB -> 20Gi */}}
          sizeLimit: {{ .quota | replace "B" "i" }}
          {{- end}}
      {{- end}}
      {{- end}}
    {{- end}}
  {{- end}}
{{- end -}}
