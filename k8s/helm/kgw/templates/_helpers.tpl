{{/*
Expand the name of the chart.
*/}}
{{- define "kgw.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kgw.fullname" -}}
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
{{- define "kgw.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kgw.labels" -}}
helm.sh/chart: {{ include "kgw.chart" . }}
{{ include "kgw.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kgw.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kgw.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "kgw.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kgw.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
configMap volume
*/}}
{{- define "kgw.configMapVolume.name" -}}
{{- $chartName := include "kgw.fullname" . -}}
{{- printf "%s-configmap-volume" $chartName }}
{{- end }}

{{/*
container probes
*/}}
{{- define "kgw.probes" -}}
livenessProbe:
  httpGet:
    path: /healthz
    port: {{ .Values.containerPorts.kgw }}
    httpHeaders:
    - name: X-Api-Key
      value: {{ .Values.kgw.healthcheckKey | sha256sum }}
  initialDelaySeconds: 15
  timeoutSeconds: 1
  periodSeconds: 15
readinessProbe:
  httpGet:
    path: /readyz
    port: {{ .Values.containerPorts.kgw }}
    httpHeaders:
    - name: X-Api-Key
      value: {{ .Values.kgw.healthcheckKey | sha256sum }}
  initialDelaySeconds: 5
  timeoutSeconds: 1
  periodSeconds: 15
{{- end }}