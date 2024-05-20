{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "numaflow.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "numaflow.labels" -}}
helm.sh/chart: {{ include "numaflow.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "server.configs.port" -}}
{{- if .Values.server.configs.insecure -}}
{{- .Values.server.configs.insecurePort}}
{{- else -}}
{{- .Values.server.configs.port }}
{{- end -}}
{{- end -}}
