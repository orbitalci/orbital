{{/* define all the token init / sidecar pieces since they'll be used everywhere  */}}
{{ define "tokenInit" }}
{{ if .Values.k8sToken.enabled }}
      initContainers:
      - name: token-grinch
        image: "{{ .Values.k8sToken.init.image }}:{{ .Values.k8sToken.init.tag }}"
        env:
        - name: VAULT_ADDR
          value: "{{.Values.vault.protocol}}://{{.Values.vault.ip}}:{{.Values.vault.port}}"
        - name: VAULT_ROLE
          value: "{{ .Values.k8sToken.vaultRole }}"
        volumeMounts:
          - name: vault-token-vol
            mountPath: "{{ .Values.k8sToken.mountPath }}"
{{ end }}
{{ end }}
{{ define "tokenSidecar" }}
{{ if .Values.k8sToken.enabled }}
        - name: token-renewer
          image: "{{ .Values.k8sToken.sidecar.image }}:{{ .Values.k8sToken.sidecar.tag }}"
          env:
          - name: VAULT_ADDR
            value: "{{.Values.vault.protocol}}://{{.Values.vault.ip}}:{{.Values.vault.port}}"
          - name: TOKEN_PATH
            value: "{{ .Values.k8sToken.mountPath }}/{{ .Values.k8sToken.tokenFileName }}"
          - name: INCREMENT_AMOUNT
          volumeMounts:
            - name: vault-token-vol
              mountPath: "{{ .Values.k8sToken.mountPath }}"
        {{ end }}
{{ end }}
{{ define "tokenVolumeSpec" }}
{{ if .Values.k8sToken.enabled }}
      - name: vault-token-vol
        emptyDir: {}
{{ end }}
{{ end }}
{{ define "tokenMountPath" }}
{{ if .Values.k8sToken.enabled }}
            - mountPath: {{ .Values.k8sToken.mountPath }}
              name: vault-token-vol
          {{ end }}
{{ end }}
{{ define "envToken" }}
{{ if not .Values.k8sToken.enabled }}
            - name: VAULT_TOKEN
              valueFrom:
                secretKeyRef:
                  name: {{.Values.vault.secretName}}
                  key: {{.Values.vault.secretKey}}
            {{ end }}
{{ end }}