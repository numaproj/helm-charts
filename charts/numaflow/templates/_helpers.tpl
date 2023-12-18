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

