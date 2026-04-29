{{/* Define namespace of chart, useful for multi-namespace deployments */}}
{{- define "error-pages.namespace" -}}
  {{- if .Values.namespaceOverride }}
    {{- .Values.namespaceOverride }}
  {{- else }}
    {{- .Release.Namespace }}
  {{- end }}
{{- end }}

{{/* Expand the name of the chart */}}
{{- define "error-pages.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "error-pages.fullname" -}}
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

{{/* Create chart name and version as used by the chart label */}}
{{- define "error-pages.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/* Common labels */}}
{{- define "error-pages.commonLabels" -}}
helm.sh/chart: {{ include "error-pages.chart" . }}
{{ include "error-pages.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/* Selector labels */}}
{{- define "error-pages.selectorLabels" -}}
app.kubernetes.io/name: {{ include "error-pages.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
