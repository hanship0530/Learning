{{- define "pacman.selectorLabels" -}}   # stetement 이름을 정의
app.kubernetes.io/name: {{ .Chart.Name}} # 해당 stetement 가 하는 일을 정의
{{- end }}
