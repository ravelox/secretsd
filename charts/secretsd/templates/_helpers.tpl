{{- define "secretsd.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- define "secretsd.fullname" -}}
{{- if .Values.fullnameOverride -}}{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}{{- else -}}{{- printf "%s" (include "secretsd.name" .) | trunc 63 | trimSuffix "-" -}}{{- end -}}
{{- end -}}
