{{/*
Expand the name of the chart.
*/}}
{{- define "node-undertaker.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "node-undertaker.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "node-undertaker.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "node-undertaker.labels" -}}
helm.sh/chart: {{ include "node-undertaker.chart" . }}
{{ include "node-undertaker.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "node-undertaker.selectorLabels" -}}
app.kubernetes.io/name: {{ include "node-undertaker.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "node-undertaker.serviceAccountName" -}}
{{- if .Values.deployment.serviceAccount.create }}
{{- default (include "node-undertaker.fullname" .) .Values.deployment.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.deployment.serviceAccount.name }}
{{- end }}
{{- end }}



{{/*##### reporter */}}

{{/*
Common labels
*/}}
{{- define "node-undertaker-reporter.labels" -}}
helm.sh/chart: {{ include "node-undertaker.chart" . }}
{{ include "node-undertaker-reporter.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "node-undertaker-reporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "node-undertaker.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}


{{/*
Create the name of the service account to use
*/}}
{{- define "node-undertaker-reporter.serviceAccountName" -}}
{{- if .Values.reporter.serviceAccount.create }}
{{- default (include "node-undertaker.fullname" .) .Values.reporter.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.reporter.serviceAccount.name }}
{{- end }}
{{- end }}
