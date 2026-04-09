{{/*
thunder-idp ComponentType spec — defined once, used in both ComponentType and ComponentRelease.
This avoids a cluster lookup and ensures the frozen snapshot always matches the chart's definition.
*/}}
{{- define "thunder.componentTypeSpec" -}}
workloadType: deployment
parameters:
  openAPIV3Schema:
    type: object
    properties:
      initial:
        type: object
        properties:
          database:
            type: object
            properties:
              config:
                type: object
                properties:
                  port: { type: string, default: "5432" }
                  sslmode: { type: string, default: "disable" }
              runtime:
                type: object
                properties:
                  port: { type: string, default: "5432" }
                  sslmode: { type: string, default: "disable" }
              user:
                type: object
                properties:
                  port: { type: string, default: "5432" }
                  sslmode: { type: string, default: "disable" }
          crypto:
            type: object
            properties:
              encryptionKey: { type: string, default: "file://repository/resources/security/crypto.key" }
      runtime:
        type: object
        properties:
          server:
            type: object
            properties:
              port: { type: integer, default: 8090 }
          jwt:
            type: object
            properties:
              issuer: { type: string, default: "" }
              validity: { type: integer, default: 3600 }
          oauth:
            type: object
            properties:
              refreshTokenValidity: { type: integer, default: 86400 }
          cache:
            type: object
            properties:
              size: { type: integer, default: 10000 }
              ttl: { type: integer, default: 3600 }
          consent:
            type: object
            properties:
              enabled: { type: boolean, default: false }
              baseUrl: { type: string, default: "http://localhost:9090/api/v1" }
          imagePullPolicy:
            type: string
            enum: [Always, IfNotPresent, Never]
            default: "IfNotPresent"
          console:
            type: object
            properties:
              clientBase: { type: string, default: "/console" }
              clientId: { type: string, default: "console" }
              scopes: { type: string, default: "[\"openid\", \"profile\", \"email\"]" }
          gate:
            type: object
            properties:
              clientBase: { type: string, default: "/gate" }
environmentConfigs:
  openAPIV3Schema:
    type: object
    properties:
      replicas:
        type: integer
        default: 1
      configDbHostname:
        type: string
        default: "<CONFIG_DB_HOST>"
      configDbName:
        type: string
        default: "<CONFIG_DB_NAME>"
      configDbUsername:
        type: string
        default: "<CONFIG_DB_USERNAME>"
      configDbPassword:
        type: string
        default: "<CONFIG_DB_PASSWORD>"
      runtimeDbHostname:
        type: string
        default: "<RUNTIME_DB_HOST>"
      runtimeDbName:
        type: string
        default: "<RUNTIME_DB_NAME>"
      runtimeDbUsername:
        type: string
        default: "<RUNTIME_DB_USERNAME>"
      runtimeDbPassword:
        type: string
        default: "<RUNTIME_DB_PASSWORD>"
      userDbHostname:
        type: string
        default: "<USER_DB_HOST>"
      userDbName:
        type: string
        default: "<USER_DB_NAME>"
      userDbUsername:
        type: string
        default: "<USER_DB_USERNAME>"
      userDbPassword:
        type: string
        default: "<USER_DB_PASSWORD>"
      serverPublicUrl:
        type: string
        default: "<SERVER_PUBLIC_URL>"
      gateClientHostname:
        type: string
        default: "<GATE_HOSTNAME>"
      gateClientPort:
        type: integer
        default: 19080
      gateClientScheme:
        type: string
        enum: [http, https]
        default: "http"
      corsAllowedOrigins:
        type: array
        items: { type: string }
        default: []
resources:
  - id: setup-job
    template:
      apiVersion: batch/v1
      kind: Job
      metadata:
        name: "${metadata.componentName}-setup"
        namespace: "${metadata.namespace}"
      spec:
        backoffLimit: 3
        ttlSecondsAfterFinished: 300
        template:
          spec:
            restartPolicy: OnFailure
            containers:
              - name: se
                image: "${workload.container.image}"
                imagePullPolicy: "${parameters.runtime.imagePullPolicy}"
                command: ["./setup.sh"]
                env:
                  - name: WITH_CONSENT
                    value: "false"
                volumeMounts:
                  - name: thunder-config
                    mountPath: /opt/thunder/repository/conf/deployment.yaml
                    subPath: deployment.yaml
            volumes:
              - name: thunder-config
                configMap:
                  name: "${metadata.componentName}-config"

  - id: thunder-config
    template:
      apiVersion: v1
      kind: ConfigMap
      metadata:
        name: "${metadata.componentName}-config"
        namespace: "${metadata.namespace}"
      data:
        deployment.yaml: |
          server:
            hostname: "0.0.0.0"
            public_url: "${environmentConfigs.serverPublicUrl}"
            port: ${parameters.runtime.server.port}
            http_only: true

          gate_client:
            hostname: "${environmentConfigs.gateClientHostname}"
            port: ${environmentConfigs.gateClientPort}
            scheme: "${environmentConfigs.gateClientScheme}"

          crypto:
            encryption:
              key: "${parameters.initial.crypto.encryptionKey}"

          database:
            config:
              type: "postgres"
              postgres:
                hostname: "${environmentConfigs.configDbHostname}"
                port: ${parameters.initial.database.config.port}
                name: "${environmentConfigs.configDbName}"
                username: "${environmentConfigs.configDbUsername}"
                password: "${environmentConfigs.configDbPassword}"
                sslmode: "${parameters.initial.database.config.sslmode}"
            runtime:
              type: "postgres"
              postgres:
                hostname: "${environmentConfigs.runtimeDbHostname}"
                port: ${parameters.initial.database.runtime.port}
                name: "${environmentConfigs.runtimeDbName}"
                username: "${environmentConfigs.runtimeDbUsername}"
                password: "${environmentConfigs.runtimeDbPassword}"
                sslmode: "${parameters.initial.database.runtime.sslmode}"
            user:
              type: "postgres"
              postgres:
                hostname: "${environmentConfigs.userDbHostname}"
                port: ${parameters.initial.database.user.port}
                name: "${environmentConfigs.userDbName}"
                username: "${environmentConfigs.userDbUsername}"
                password: "${environmentConfigs.userDbPassword}"
                sslmode: "${parameters.initial.database.user.sslmode}"

          cache:
            disabled: false
            type: "memory"
            size: ${parameters.runtime.cache.size}
            ttl: ${parameters.runtime.cache.ttl}
            eviction_policy: "lru"
            cleanup_interval: 600

          jwt:
            issuer: "${parameters.runtime.jwt.issuer}"
            validity_period: ${parameters.runtime.jwt.validity}

          oauth:
            refresh_token:
              renew_on_grant: true
              validity_period: ${parameters.runtime.oauth.refreshTokenValidity}

          flow:
            max_version_history: 3
            auto_infer_registration: true

          cors:
            allowed_origins: ${environmentConfigs.corsAllowedOrigins}

          consent:
            enabled: ${parameters.runtime.consent.enabled}
            base_url: "${parameters.runtime.consent.baseUrl}"

  - id: gate-config
    template:
      apiVersion: v1
      kind: ConfigMap
      metadata:
        name: "${metadata.componentName}-gate-config"
        namespace: "${metadata.namespace}"
      data:
        config.js: |
          window.__THUNDER_RUNTIME_CONFIG__ = {
            client: {
              base: "${parameters.runtime.gate.clientBase}",
            },
            server: {
              public_url: "${environmentConfigs.serverPublicUrl}",
            },
          };

  - id: console-config
    template:
      apiVersion: v1
      kind: ConfigMap
      metadata:
        name: "${metadata.componentName}-console-config"
        namespace: "${metadata.namespace}"
      data:
        config.js: |
          window.__THUNDER_RUNTIME_CONFIG__ = {
            client: {
              base: "${parameters.runtime.console.clientBase}",
              client_id: "${parameters.runtime.console.clientId}",
              scopes: ${parameters.runtime.console.scopes},
            },
            server: {
              public_url: "${environmentConfigs.serverPublicUrl}",
            },
          };

  - id: deployment
    template:
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: "${metadata.componentName}"
        namespace: "${metadata.namespace}"
      spec:
        replicas: ${environmentConfigs.replicas}
        selector:
          matchLabels:
            app: "${metadata.componentName}"
        template:
          metadata:
            labels:
              app: "${metadata.componentName}"
          spec:
            containers:
              - name: thunder
                image: "${workload.container.image}"
                imagePullPolicy: "${parameters.runtime.imagePullPolicy}"
                command:
                  - /opt/thunder/thunder
                ports:
                  - containerPort: ${parameters.runtime.server.port}
                volumeMounts:
                  - name: thunder-config
                    mountPath: /opt/thunder/repository/conf/deployment.yaml
                    subPath: deployment.yaml
                  - name: gate-config
                    mountPath: /opt/thunder/apps/gate/config.js
                    subPath: config.js
                  - name: console-config
                    mountPath: /opt/thunder/apps/console/config.js
                    subPath: config.js
            volumes:
              - name: thunder-config
                configMap:
                  name: "${metadata.componentName}-config"
              - name: gate-config
                configMap:
                  name: "${metadata.componentName}-gate-config"
              - name: console-config
                configMap:
                  name: "${metadata.componentName}-console-config"

  - id: service
    template:
      apiVersion: v1
      kind: Service
      metadata:
        name: "${metadata.componentName}"
        namespace: "${metadata.namespace}"
      spec:
        type: ClusterIP
        selector:
          app: "${metadata.componentName}"
        ports:
          - port: ${parameters.runtime.server.port}
            targetPort: ${parameters.runtime.server.port}
            protocol: TCP
            name: http

  - id: httproute
    includeWhen: "${has(gateway.ingress) && has(gateway.ingress.external)}"
    template:
      apiVersion: gateway.networking.k8s.io/v1
      kind: HTTPRoute
      metadata:
        name: "${metadata.componentName}"
        namespace: "${metadata.namespace}"
      spec:
        parentRefs:
          - name: "${gateway.ingress.external.name}"
            namespace: "${gateway.ingress.external.namespace}"
        hostnames:
          - "${metadata.componentName}-${metadata.environmentName}.${gateway.ingress.external.http.host}"
        rules:
          - matches:
              - path:
                  type: PathPrefix
                  value: /
            backendRefs:
              - name: "${metadata.componentName}"
                port: ${parameters.runtime.server.port}
{{- end }}
