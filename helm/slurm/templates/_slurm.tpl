{{- /*
SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
SPDX-License-Identifier: Apache-2.0
*/}}

{{/*
Define auth/slurm secret ref name
*/}}
{{- define "slurm.authSlurmRef.name" -}}
{{- if .Values.slurmKey.secretRef.name }}
{{- .Values.slurmKey.secretRef.name }}
{{- else }}
{{- printf "%s-auth-slurm" (include "slurm.fullname" .) -}}
{{- end }}
{{- end }}

{{/*
Define auth/slurm secret ref key
*/}}
{{- define "slurm.authSlurmRef.key" -}}
{{- if .Values.slurmKey.secretRef.key }}
{{- .Values.slurmKey.secretRef.key }}
{{- else }}
{{- print "slurm.key" -}}
{{- end }}
{{- end }}

{{/*
Define auth/jwt secret ref name
*/}}
{{- define "slurm.authJwtRef.name" -}}
{{- if .Values.jwtKey.secretRef.name }}
{{- .Values.jwtKey.secretRef.name }}
{{- else }}
{{- printf "%s-auth-jwt" (include "slurm.fullname" .) -}}
{{- end }}
{{- end }}

{{/*
Define auth/jwt secret ref key
*/}}
{{- define "slurm.authJwtRef.key" -}}
{{- if .Values.jwtKey.secretRef.key }}
{{- .Values.jwtKey.secretRef.key }}
{{- else }}
{{- print "jwt.key" -}}
{{- end }}
{{- end }}

{{/*
Define JWKS configMap ref name
*/}}
{{- define "slurm.authJwksRef.name" -}}
{{- if .Values.jwksKeys.configMapRef.name }}
{{- .Values.jwksKeys.configMapRef.name }}
{{- else }}
{{- printf "%s-auth-jwks" (include "slurm.fullname" .) -}}
{{- end }}
{{- end }}

{{/*
Define JWKS configMap ref key
*/}}
{{- define "slurm.authJwksRef.key" -}}
{{- if .Values.jwksKeys.configMapRef.key }}
{{- .Values.jwksKeys.configMapRef.key }}
{{- else }}
{{- print "jwks.json" -}}
{{- end }}
{{- end }}
