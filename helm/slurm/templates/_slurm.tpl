{{- /*
SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
SPDX-License-Identifier: Apache-2.0
*/}}

{{/*
Define slurm User
*/}}
{{- define "slurm.user" -}}
{{- print "slurm" -}}
{{- end }}

{{/*
Define slurm UID
*/}}
{{- define "slurm.uid" -}}
{{- print "401" -}}
{{- end }}

{{/*
Determine authcred image repository
*/}}
{{- define "slurm.authcred.image.repository" -}}
{{- .Values.authcred.image.repository | default (printf "%s/sackd" (include "slurm.image.repository" .)) -}}
{{- end }}

{{/*
Define authcred image tag
*/}}
{{- define "slurm.authcred.image.tag" -}}
{{- .Values.authcred.image.tag | default (include "slurm.image.tag" .) -}}
{{- end }}

{{/*
Determine authcred image reference (repo:tag)
*/}}
{{- define "slurm.authcred.imageRef" -}}
{{- printf "%s:%s" (include "slurm.authcred.image.repository" .) (include "slurm.authcred.image.tag" .) | quote -}}
{{- end }}

{{/*
Define controller name
*/}}
{{- define "slurm.controller.name" -}}
{{- printf "%s-controller" (.Release.Name) -}}
{{- end }}

{{/*
Define controller port
*/}}
{{- define "slurm.controller.port" -}}
{{- print "6817" -}}
{{- end }}

{{/*
Define controller labels
*/}}
{{- define "slurm.controller.labels" -}}
app.kubernetes.io/component: controller
{{ include "slurm.controller.selectorLabels" . }}
{{ include "slurm.labels" . }}
{{- end }}

{{/*
Define controller selectorLabels
*/}}
{{- define "slurm.controller.selectorLabels" -}}
app.kubernetes.io/name: slurmctld
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Determine controller image repository
*/}}
{{- define "slurm.controller.image.repository" -}}
{{- .Values.controller.image.repository | default (printf "%s/slurmctld" (include "slurm.image.repository" .)) -}}
{{- end }}

{{/*
Define controller image tag
*/}}
{{- define "slurm.controller.image.tag" -}}
{{- .Values.controller.image.tag | default (include "slurm.image.tag" .) -}}
{{- end }}

{{/*
Determine controller image reference (repo:tag)
*/}}
{{- define "slurm.controller.imageRef" -}}
{{- printf "%s:%s" (include "slurm.controller.image.repository" .) (include "slurm.controller.image.tag" .) | quote -}}
{{- end }}

{{/*
Define controller state save name
*/}}
{{- define "slurm.controller.statesave.name" -}}
{{- print "statesave" -}}
{{- end }}

{{/*
Define controller state save path
*/}}
{{- define "slurm.controller.statesavePath" -}}
{{- print "/var/spool/slurmctld" -}}
{{- end }}

{{/*
Define controller log file
*/}}
{{- define "slurm.controller.logFile" -}}
{{- print "/var/log/slurm/slurmctld.log" -}}
{{- end }}

{{/*
Define accounting name
*/}}
{{- define "slurm.accounting.name" -}}
{{- printf "%s-accounting" (.Release.Name) -}}
{{- end }}

{{/*
Determine accounting port
*/}}
{{- define "slurm.accounting.port" -}}
{{- print "6819" -}}
{{- end }}

{{/*
Define accounting labels
*/}}
{{- define "slurm.accounting.labels" -}}
app.kubernetes.io/component: accounting
{{ include "slurm.accounting.selectorLabels" . }}
{{ include "slurm.labels" . }}
{{- end }}

{{/*
Define accounting selectorLabels
*/}}
{{- define "slurm.accounting.selectorLabels" -}}
app.kubernetes.io/name: slurmdbd
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Determine accounting image repository
*/}}
{{- define "slurm.accounting.image.repository" -}}
{{ .Values.accounting.image.repository | default (printf "%s/slurmdbd" (include "slurm.image.repository" .)) }}
{{- end }}

{{/*
Determine accounting image tag
*/}}
{{- define "slurm.accounting.image.tag" -}}
{{- .Values.accounting.image.tag | default (include "slurm.image.tag" .) -}}
{{- end }}

{{/*
Determine accounting image reference (repo:tag)
*/}}
{{- define "slurm.accounting.imageRef" -}}
{{- printf "%s:%s" (include "slurm.accounting.image.repository" .) (include "slurm.accounting.image.tag" .) | quote -}}
{{- end }}

{{/*
Define slurm accounting config
*/}}
{{- define "slurm.accounting.config.name" -}}
{{- printf "%s-accounting" (.Release.Name) -}}
{{- end }}

{{/*
Define slurm accounting database secret
*/}}
{{- define "slurm.accounting.secretName" -}}
{{- if and .Values.accounting.external.enabled .Values.accounting.external.secretName }}
  {{- print .Values.accounting.external.secretName }}
{{- else if and .Values.accounting.external.enabled }}
  {{- printf "%s-database" (include "slurm.accounting.name" .) }}
{{- else }}
  {{- template "mariadb.secretName" .Subcharts.mariadb }}
{{- end }}
{{- end }}

{{/*
Determine compute image repository
*/}}
{{- define "slurm.compute.image.repository" -}}
{{- .Values.compute.image.repository | default (printf "%s/slurmd" (include "slurm.image.repository" .)) -}}
{{- end }}

{{/*
Define image tag
*/}}
{{- define "slurm.compute.image.tag" -}}
{{- .Values.compute.image.tag | default (include "slurm.image.tag" .) -}}
{{- end }}

{{/*
Define compute name
*/}}
{{- define "slurm.compute.name" -}}
{{- printf "%s-compute" (.Release.Name) -}}
{{- end }}

{{/*
Define compute port
*/}}
{{- define "slurm.compute.port" -}}
{{- print "6818" -}}
{{- end }}

{{/*
Define compute spool directory
*/}}
{{- define "slurm.compute.spoolDir" -}}
{{- print "/var/spool/slurmd" -}}
{{- end }}

{{/*
Define compute labels
*/}}
{{- define "slurm.compute.labels" -}}
app.kubernetes.io/component: compute
{{ include "slurm.compute.selectorLabels" . }}
{{ include "slurm.labels" . }}
{{- end }}

{{/*
Define compute selectorLabels
*/}}
{{- define "slurm.compute.selectorLabels" -}}
app.kubernetes.io/name: slurmd
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Define slurmd capabilities
*/}}
{{- define "slurm.compute.capabilities" -}}
- BPF
- NET_ADMIN
- SYS_ADMIN
- SYS_NICE
{{- end }}

{{/*
Define compute log file
*/}}
{{- define "slurm.compute.logFile" -}}
{{- print "/var/log/slurm/slurmd.log" -}}
{{- end }}

{{/*
Returns the parsed resource limits for POD_CPUS.
*/}}
{{- define "slurm.compute.podCpus" -}}
{{- $out := 0 -}}
{{- with .resources }}{{- with .limits }}{{- with .cpu }}
  {{- $out = include "resource-quantity" . | float64 | ceil | int -}}
{{- end }}{{- end }}{{- end }}
{{- print $out -}}
{{- end -}}

{{/*
Returns the parsed resource limits for POD_MEMORY, in Megabytes.
*/}}
{{- define "slurm.compute.podMemory" -}}
{{- $out := 0 -}}
{{- with .resources }}{{- with .limits }}{{- with .memory }}
  {{- $megabytes := (include "resource-quantity" "1M") | float64 -}}
  {{- $out = divf (include "resource-quantity" . | float64) $megabytes | ceil | int -}}
{{- end }}{{- end }}{{- end }}
{{- print $out -}}
{{- end -}}

{{/*
Returns the --conf line given the NodeSet.
Ref: https://slurm.schedmd.com/slurmd.html#OPT_conf-%3Cnode-parameters%3E
Ref: https://slurm.schedmd.com/slurm.conf.html#SECTION_NODE-CONFIGURATION
*/}}
{{- define "slurm.compute.conf" -}}
{{- $confList := list -}}
{{- $featureList := list .name -}}
{{- $featureKeyList := list "features" "feature" -}}
{{- range $key, $val := .nodeConfig -}}
  {{- if has (lower $key) $featureKeyList -}}
    {{- if typeIs "string" $val -}}
      {{- $featureList = concat $featureList (list $val) -}}
    {{- else -}}
      {{- $featureList = concat $featureList $val -}}
    {{- end -}}
  {{- else -}}
    {{- if $val -}}
      {{- $item := printf "%s=%s" $key ($val | join ",") -}}
      {{- $confList = append $confList $item -}}
    {{- end -}}
  {{- end -}}
{{- end -}}
{{- $item := printf "Features=%s" (join "," $featureList) -}}
{{- $confList = prepend $confList $item -}}
{{- printf "'%s'" (join " " $confList) | quote -}}
{{- end -}}

{{/*
Determine login image repository
*/}}
{{- define "slurm.login.image.repository" -}}
{{- .Values.login.image.repository | default (printf "%s/sackd" (include "slurm.image.repository" .)) -}}
{{- end }}

{{/*
Define login image tag
*/}}
{{- define "slurm.login.image.tag" -}}
{{- .Values.login.image.tag | default (include "slurm.image.tag" .) -}}
{{- end }}

{{/*
Determine login image reference (repo:tag)
*/}}
{{- define "slurm.login.imageRef" -}}
{{- printf "%s:%s" (include "slurm.login.image.repository" .) (include "slurm.login.image.tag" .) | quote -}}
{{- end }}

{{/*
Define restapi name
*/}}
{{- define "slurm.restapi.name" -}}
{{- printf "%s-restapi" (.Release.Name) -}}
{{- end }}

{{/*
Define restapi port
*/}}
{{- define "slurm.restapi.port" -}}
{{- print "6820" -}}
{{- end }}

{{/*
Define restapi labels
*/}}
{{- define "slurm.restapi.labels" -}}
app.kubernetes.io/component: restapi
{{ include "slurm.restapi.selectorLabels" . }}
{{ include "slurm.labels" . }}
{{- end }}

{{/*
Define restapi selectorLabels
*/}}
{{- define "slurm.restapi.selectorLabels" -}}
app.kubernetes.io/name: slurmrestd
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Determine restapi image repository
*/}}
{{- define "slurm.restapi.image.repository" -}}
{{- .Values.restapi.image.repository | default (printf "%s/slurmrestd" (include "slurm.image.repository" .)) -}}
{{- end }}

{{/*
Determine restapi image tag
*/}}
{{- define "slurm.restapi.image.tag" -}}
{{- .Values.restapi.image.tag | default (include "slurm.image.tag" .) -}}
{{- end }}

{{/*
Determine restapi image reference (repo:tag)
*/}}
{{- define "slurm.restapi.imageRef" -}}
{{- printf "%s:%s" (include "slurm.restapi.image.repository" .) (include "slurm.restapi.image.tag" .) | quote -}}
{{- end }}

{{/*
Define cluster name
*/}}
{{- define "slurm.cluster.name" -}}
{{- printf "%s" (include "slurm.name" .) -}}
{{- end }}

{{/*
Define cluster labels
*/}}
{{- define "slurm.cluster.labels" -}}
app.kubernetes.io/component: cluster
{{ include "slurm.cluster.selectorLabels" . }}
{{ include "slurm.labels" . }}
{{- end }}

{{/*
Define cluster selectorLabels
*/}}
{{- define "slurm.cluster.selectorLabels" -}}
app.kubernetes.io/name: cluster
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Define cluster secret name
*/}}
{{- define "slurm.cluster.secretName" -}}
{{- printf "%s-token-%s" .Release.Name (include "slurm.user" .) -}}
{{- end }}

{{/*
Define login name
*/}}
{{- define "slurm.login.name" -}}
{{ printf "%s-login" .Release.Name }}
{{- end }}

{{/*
Define login labels
*/}}
{{- define "slurm.login.labels" -}}
app.kubernetes.io/component: login
{{ include "slurm.login.selectorLabels" . }}
{{ include "slurm.labels" . }}
{{- end }}

{{/*
Define login selectorLabels
*/}}
{{- define "slurm.login.selectorLabels" -}}
app.kubernetes.io/name: login
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Define token name
*/}}
{{- define "slurm.token.name" -}}
{{ printf "%s-token-create" .Release.Name }}
{{- end }}

{{/*
Define token labels
*/}}
{{- define "slurm.token.labels" -}}
app.kubernetes.io/component: token
{{ include "slurm.token.selectorLabels" . }}
{{ include "slurm.labels" . }}
{{- end }}

{{/*
Define token selectorLabels
*/}}
{{- define "slurm.token.selectorLabels" -}}
app.kubernetes.io/name: token
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Define slurm auth secret name
*/}}
{{- define "slurm.auth.secretName" -}}
{{- if ((.Values.slurm).auth).secretName -}}
  {{- printf "%s" (tpl .Values.slurm.auth.secretName $) -}}
{{- else -}}
  {{- printf "%s-auth-key" (.Release.Name) -}}
{{- end -}}
{{- end }}

{{/*
Define slurm securityContext
*/}}
{{- define "slurm.securityContext" -}}
runAsNonRoot: true
runAsUser: {{ include "slurm.uid" . }}
runAsGroup: {{ include "slurm.uid" . }}
{{- end }}

{{/*
Define slurm mountPath
*/}}
{{- define "slurm.mountPath" -}}
{{- print "/etc/slurm" -}}
{{- end }}

{{/*
Define slurm jwt hs256 secret name
*/}}
{{- define "slurm.jwt.hs256.secretName" -}}
{{- if ((.Values.jwt).hs256).secretName -}}
  {{- printf "%s" (tpl .Values.jwt.hs256.secretName $) -}}
{{- else -}}
  {{- printf "%s-jwt-key" (.Release.Name) -}}
{{- end -}}
{{- end }}

{{/*
Define jwt hs256 key path
*/}}
{{- define "slurm.jwt.hs256.fullPath" -}}
{{- print "/etc/slurm/jwt_hs256.key" -}}
{{- end }}

{{/*
Define Slurm config name
*/}}
{{- define "slurm.configMapName" -}}
{{- printf "%s-config" (.Release.Name) -}}
{{- end }}

{{/*
Define user auth name
*/}}
{{- define "slurm.userauth.name" -}}
{{- printf "%s-userauth" (.Release.Name) -}}
{{- end }}

{{/*
Common volumes
*/}}
{{- define "slurm.volumes" -}}
- name: etc-slurm
  emptyDir:
    medium: Memory
- name: run
  emptyDir: {}
{{- end }}

{{/*
Common volumeMounts
*/}}
{{- define "slurm.volumeMounts" -}}
- name: etc-slurm
  mountPath: /etc/slurm
- name: run
  mountPath: /run
{{- end }}

{{/*
Common dnsConfig
*/}}
{{- define "slurm.dnsConfig" -}}
searches:
  - {{ include "slurm.controller.name" . -}}.{{- .Release.Namespace -}}.svc.cluster.local
  - {{ include "slurm.compute.name" . -}}.{{- .Release.Namespace -}}.svc.cluster.local
{{- end }}

{{/*
Define slurm initContainers volumeMounts
*/}}
{{- define "slurm.init.volumeMounts" -}}
- name: slurm-config
  mountPath: /mnt/slurm
- name: etc-slurm
  mountPath: /mnt/etc/slurm
{{- end }}
