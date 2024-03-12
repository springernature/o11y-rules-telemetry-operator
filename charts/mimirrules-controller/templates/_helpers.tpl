{{/*
Expand the name of the chart.
*/}}
{{- define "chart.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "chart.fullname" -}}
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
{{- define "chart.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "chart.labels" -}}
helm.sh/chart: {{ include "chart.chart" . }}
{{ include "chart.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "chart.selectorLabels" -}}
app.kubernetes.io/name: {{ include "chart.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "chart.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "chart.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "mimirrules" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the base name of the clusterroles to use
*/}}
{{- define "chart.crdClusterRolesBaseName" -}}
{{- if .Values.crdClusterRoles.create }}
{{- default (include "chart.fullname" .) .Values.crdClusterRoles.name }}
{{- else }}
{{- default "mimirrules" .Values.crdClusterRoles.name }}
{{- end }}
{{- end }}

{{/*
Create the name for the controller clusterrole
*/}}
{{- define "chart.controllerClusterRoleName" -}}
{{- if .Values.controllerClusterRole.create }}
{{- default (include "chart.fullname" .) .Values.controllerClusterRole.name }}
{{- else }}
{{- default "crmimirrules" .Values.controllerClusterRole.name }}
{{- end }}
{{- end }}

{{/*
Create the name for the auth-proxy clusterrole
*/}}
{{- define "chart.authProxyClusterRoleName" -}}
{{- if .Values.authProxyClusterRole.create }}
{{- default (include "chart.fullname" .) .Values.authProxyClusterRole.name }}
{{- else }}
{{- default "authproxymimirrules" .Values.authProxyClusterRole.name }}
{{- end }}
{{- end }}