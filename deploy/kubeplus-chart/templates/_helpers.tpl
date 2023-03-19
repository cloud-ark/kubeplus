{{- define "webhooktlsGetterImage" -}}
{{- $registryName := .Values.webhooktlsGetter.image.registry -}}
{{- $imageName := .Values.webhooktlsGetter.image.repository -}}
{{- $tag := .Values.webhooktlsGetter.image.tag -}}
{{- printf "%s/%s:%s" $registryName $imageName $tag -}}
{{- end -}}

{{- define "kubeconfiggeneratorImage" -}}
{{- $registryName := .Values.kubeconfiggenerator.image.registry -}}
{{- $imageName := .Values.kubeconfiggenerator.image.repository -}}
{{- $tag := .Values.kubeconfiggenerator.image.tag -}}
{{- printf "%s/%s:%s" $registryName $imageName $tag -}}
{{- end -}}

{{- define "mutatingAdmissionWebhookImage" -}}
{{- $registryName := .Values.mutatingAdmissionWebhook.image.registry -}}
{{- $imageName := .Values.mutatingAdmissionWebhook.image.repository -}}
{{- $tag := .Values.mutatingAdmissionWebhook.image.tag -}}
{{- printf "%s/%s:%s" $registryName $imageName $tag -}}
{{- end -}}

{{- define "platformOperatorImage" -}}
{{- $registryName := .Values.platformOperator.image.registry -}}
{{- $imageName := .Values.platformOperator.image.repository -}}
{{- $tag := .Values.platformOperator.image.tag -}}
{{- printf "%s/%s:%s" $registryName $imageName $tag -}}
{{- end -}}


{{- define "consumeruiImage" -}}
{{- $registryName := .Values.consumerui.image.registry -}}
{{- $imageName := .Values.consumerui.image.repository -}}
{{- $tag := .Values.consumerui.image.tag -}}
{{- printf "%s/%s:%s" $registryName $imageName $tag -}}
{{- end -}}

{{- define "helmerImage" -}}
{{- $registryName := .Values.helmer.image.registry -}}
{{- $imageName := .Values.helmer.image.repository -}}
{{- $tag := .Values.helmer.image.tag -}}
{{- printf "%s/%s:%s" $registryName $imageName $tag -}}
{{- end -}}

{{- define "cleanupKubeplusComponentsImage" -}}
{{- $registryName := .Values.cleanupKubeplusComponents.image.registry -}}
{{- $imageName := .Values.cleanupKubeplusComponents.image.repository -}}
{{- $tag := .Values.cleanupKubeplusComponents.image.tag -}}
{{- printf "%s/%s:%s" $registryName $imageName $tag -}}
{{- end -}}