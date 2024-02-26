{{/*
Expand the name of the chart.
*/}}
{{- define "kwild.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kwild.fullname" -}}
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
{{- define "kwild.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kwild.labels" -}}
helm.sh/chart: {{ include "kwild.chart" . }}
{{ include "kwild.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kwild.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kwild.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service config to use
*/}}
{{- define "kwild.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kwild.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
initContainers for hasura depencency
*/}}
{{- define "kwild.initContainers.hasura" -}}
- name: init-wait-hasura-service
  image: busybox:1.35.0
  command:
    - sh
    - -c
    - |
      {{- $service := printf "%s-hasura" .Release.Name }}
      echo try resolve dns name "{{ $service }}"
      until nslookup {{ $service }}.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local
      do
        echo waiting for {{ $service }}
        sleep 2
      done
- name: init-wait-hasura
  image: busybox:1.35.0
  envFrom:
  - configMapRef:
      name: {{ include "kwild.fullname" . }}
  command:
    - sh
    - -c
    - |
      {{- $host := printf "%s-hasura" .Release.Name }}
      {{- $port := .Values.hasura.service.ports.hasura }}
      echo scan {{ $host }}:{{ $port }}
      for i in $(seq 1 300);
      do
        nc -zvw1 {{ $host}} {{ $port }} && exit 0 || sleep 3
      done
      exit 1
{{- end }}

{{/*
container probes
*/}}
{{- define "kwild.probes" -}}
livenessProbe:
  grpc:
    port: {{ .Values.containerPorts.kwild }}
  initialDelaySeconds: {{ .Values.kwildApp.probes.healthInitialDelaySeconds }}
  timeoutSeconds: 1
  periodSeconds: 15
readinessProbe:
  grpc:
    port: {{ .Values.containerPorts.kwild }}
  initialDelaySeconds: 5
  timeoutSeconds: 1
  periodSeconds: 15
{{- end }}