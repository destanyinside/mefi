{{/*
Pod template used in Deployment
*/}}
{{- define "mefi.podTemplate" -}}
metadata:
  labels:
    {{- include "mefi.selectorLabels" . | nindent 4 }}
    {{- with .Values.podLabels }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
  annotations:
    checksum/secret: {{ include (print $.Template.BasePath "/secretConfig.yaml") . | sha256sum }}
    {{- with .Values.podAnnotations }}
    {{- toYaml . | nindent 8 }}
    {{- end }}
spec:
  serviceAccountName: {{ include "mefi.serviceAccountName" . }}
  {{- with .Values.hostNetwork }}
  hostNetwork: {{ . }}
  {{- end }}
  {{- with .Values.priorityClassName }}
  priorityClassName: {{ . }}
  {{- end }}
  {{- with .Values.initContainer }}
  initContainers:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.imagePullSecrets }}
  imagePullSecrets:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.hostAliases }}
  hostAliases:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  securityContext:
    {{- toYaml .Values.podSecurityContext | nindent 4 }}
  containers:
    - name: mefi
      image: "{{ .Values.image.registry }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
      imagePullPolicy: {{ .Values.image.pullPolicy }}
      args:
        - "--config=/etc/mefi/config.yaml"
        {{- with .Values.extraArgs }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
      volumeMounts:
        - name: config
          mountPath: /etc/mefi
        {{- with .Values.extraVolumeMounts }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
      env:
      {{- with .Values.extraEnv }}
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with .Values.extraEnvFrom }}
      envFrom:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      ports:
        - name: http-metrics
          containerPort: {{ .Values.serverPort }}
          protocol: TCP
      securityContext:
        {{- toYaml .Values.containerSecurityContext | nindent 8 }}
      {{- with .Values.livenessProbe }}
      livenessProbe:
        {{- tpl (toYaml .) $ | nindent 8 }}
      {{- end }}
      {{- with .Values.readinessProbe }}
      readinessProbe:
        {{- tpl (toYaml .) $ | nindent 8 }}
      {{- end }}
      {{- with .Values.resources }}
      resources:
        {{- toYaml . | nindent 8 }}
      {{- end }}
  {{- with .Values.affinity }}
  affinity:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.nodeSelector }}
  nodeSelector:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.tolerations }}
  tolerations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  volumes:
    - name: config
      secret:
        secretName: {{ include "mefi.fullname" . }}
    {{- with .Values.extraVolumes }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
{{- end }}