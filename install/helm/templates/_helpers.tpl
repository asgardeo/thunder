{{/*
Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).

WSO2 LLC. licenses this file to you under the Apache License,
Version 2.0 (the "License"); you may not use this file except
in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/}}

{{/*
Expand the name of the chart.
*/}}

{{- define "thunder.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "thunder.fullname" -}}
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
{{- define "thunder.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "thunder.labels" -}}
helm.sh/chart: {{ include "thunder.chart" . }}
{{ include "thunder.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "thunder.selectorLabels" -}}
app.kubernetes.io/name: {{ include "thunder.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "thunder.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "thunder.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Check if auto-generated database credentials Secret should be included in checksum annotation.
Returns true if any database password is set without a passwordRef.key.
This is used to trigger pod restarts when auto-generated Secrets change.
*/}}
{{- define "thunder.shouldIncludeSecretChecksum" -}}
{{- $configuration := default dict .Values.configuration -}}
{{- $database := default dict $configuration.database -}}
{{- $identity := default dict $database.identity -}}
{{- $runtime := default dict $database.runtime -}}
{{- $user := default dict $database.user -}}
{{- if or (and $identity.password (not (default dict $identity.passwordRef).key)) (and $runtime.password (not (default dict $runtime.passwordRef).key)) (and $user.password (not (default dict $user.passwordRef).key)) }}true{{- end }}
{{- end }}

{{/*
Generate database password environment variable definitions for both deployment and setup job.
Injects DB_IDENTITY_PASSWORD, DB_RUNTIME_PASSWORD, and DB_USER_PASSWORD from either auto-generated or external Secrets.
*/}}
{{- define "thunder.databasePasswordEnvVars" -}}
{{- $defaultDbSecretName := printf "%s-db-credentials" (include "thunder.fullname" .) -}}
{{- $configuration := default dict .Values.configuration -}}
{{- $database := default dict $configuration.database -}}
{{- $identity := default dict $database.identity -}}
{{- $runtime := default dict $database.runtime -}}
{{- $user := default dict $database.user -}}
{{- $identityPasswordRef := default dict $identity.passwordRef -}}
{{- $runtimePasswordRef := default dict $runtime.passwordRef -}}
{{- $userPasswordRef := default dict $user.passwordRef -}}
{{- if or $identity.password $identityPasswordRef.key }}
- name: DB_IDENTITY_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ if $identityPasswordRef.key }}{{ $identityPasswordRef.name | default $defaultDbSecretName }}{{ else }}{{ $defaultDbSecretName }}{{ end }}
      key: {{ $identityPasswordRef.key | default "identity-db-password" }}
{{- end }}
{{- if or $runtime.password $runtimePasswordRef.key }}
- name: DB_RUNTIME_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ if $runtimePasswordRef.key }}{{ $runtimePasswordRef.name | default $defaultDbSecretName }}{{ else }}{{ $defaultDbSecretName }}{{ end }}
      key: {{ $runtimePasswordRef.key | default "runtime-db-password" }}
{{- end }}
{{- if or $user.password $userPasswordRef.key }}
- name: DB_USER_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ if $userPasswordRef.key }}{{ $userPasswordRef.name | default $defaultDbSecretName }}{{ else }}{{ $defaultDbSecretName }}{{ end }}
      key: {{ $userPasswordRef.key | default "user-db-password" }}
{{- end }}
{{- end }}
