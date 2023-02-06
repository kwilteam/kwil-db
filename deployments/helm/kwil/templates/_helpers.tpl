{{/*
Expand the name of the chart.
*/}}
{{- define "kwil.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kwil.fullname" -}}
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
{{- define "kwil.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kwil.labels" -}}
helm.sh/chart: {{ include "kwil.chart" . }}
{{ include "kwil.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "kwil.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kwil.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service config to use
*/}}
{{- define "kwil.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "kwil.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
configMap volume
*/}}
{{- define "kwil.configMapVolume.name" -}}
{{- $chartName := include "kwil.fullname" . -}}
{{- printf "%s-configmap-volume" $chartName }}
{{- end }}

{{/*
initContainers for kwild depencency
*/}}
{{- define "kwil.initContainers.kwild" -}}
- name: init-wait-kwild-service
  image: busybox:1.35.0
  command:
    - sh
    - -c
    - |
      {{- $service := printf "%s-kwild" .Release.Name }}
      echo try resolve dns name "{{ $service }}"
      until nslookup {{ $service }}.$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace).svc.cluster.local
      do
        echo waiting for {{ $service }}
        sleep 2
      done
- name: init-wait-kwild
  image: busybox:1.35.0
  envFrom:
  - configMapRef:
      name: {{ include "kwil.fullname" . }}
  command:
    - sh
    - -c
    - |
      {{- $host := printf "%s-kwild" .Release.Name }}
      {{- $port := .Values.kwild.service.ports.kwild }}
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
{{- define "kwil.probes" -}}
livenessProbe:
  httpGet:
    path: /healthz
    port: {{ .Values.containerPorts.kwil }}
    httpHeaders:
    - name: X-Api-Key
      value: {{ .Values.kwil.server.healthcheckKey | sha256sum }}
  initialDelaySeconds: 15
  timeoutSeconds: 1
  periodSeconds: 15
readinessProbe:
  httpGet:
    path: /readyz
    port: {{ .Values.containerPorts.kwil }}
    httpHeaders:
    - name: X-Api-Key
      value: {{ .Values.kwil.server.healthcheckKey | sha256sum }}
  initialDelaySeconds: 5
  timeoutSeconds: 1
  periodSeconds: 15
{{- end }}