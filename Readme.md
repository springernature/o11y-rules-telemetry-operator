# Helm repository for Mimir Rules Kubernetes controller

K8S controller to manage Mimir Alerting and Recording rules with dynamic tenants, based on namespace name or annotations.

## Adding Helm repository

[Helm](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

Once Helm is set up properly, add the repo as follows:

```console
helm repo add o11y-rules-telemetry-operator https://springernature.github.io/o11y-rules-telemetry-operator
helm repo update
```

You can then run `helm search repo o11y-rules-telemetry-operator` to see the charts.

## Using helm chart

To install the app chart in the Kubernetes cluster.

```console
helm upgrade --install mimirrules o11y-rules-telemetry-operator/mimirrules
```

This will install CRD `MimirRules`. Take this into account for CRD updates or when deleting the controller, those operations will require manual steps with `kubectl delete/apply`.


### Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| fullnameOverride | string | `nil` | Overrides the chart's computed fullname |
| nameOverride | string | `nil` | Overrides the chart's name |
| logger.develMode | bool | `true` | Controller logging settings, Development Mode defaults(encoder=consoleEncoder,logLevel=Debug,stackTraceLevel=Warn). Production Mode defaults(encoder=jsonEncoder,logLevel=Info,stackTraceLevel=Error) |
| logger.level | string | `info` | Logging level, one of 'debug', 'info', 'error' |
| logger.timeEncoding | string | `iso8601` | Time encoding (one of 'epoch', 'millis', 'nano', 'iso8601', 'rfc3339' or 'rfc3339nano') |
| config.mimirAPI | string | `nil` | Mimir URL API with ruler api enabled. |
| config.tenantAnnotation | string | `telemetry.springernature.com/o11y-tenant` | Annotation to define tenant instead of namespace. |
| config.CRLabelSelector | string | `nil` | Label selector to filter specific CR for this controller |
| affinity | string | Hard node and soft zone anti-affinity | Affinity for controller pods. Passed through `tpl` and, thus, to be configured as string |
| autoscaling.enabled | bool | `false` | Enable autoscaling for the controller |
| autoscaling.maxReplicas | int | `3` | Maximum autoscaling replicas for the controller |
| autoscaling.minReplicas | int | `1` | Minimum autoscaling replicas for the controller |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Target CPU utilisation percentage for the controller |
| autoscaling.targetMemoryUtilizationPercentage | string | `nil` | Target memory utilisation percentage for the controller |
| volumeMounts | list | `[]` | Volume mounts to add to the controller pods |
| volumes | list | `[]` | Volumes to add to the controller pods |
| image.repository | string | `ghcr.io/springernature/o11y-rules-telemetry-operator/mimirrules` | Docker image repository for the controller image. |
| image.tag | string | `nil` | Docker image tag for the controller image. |
| nodeSelector | object | `{}` | Node selector for controller pods |
| podAnnotations | object | `{}` | Annotations for controller pods |
| podLabels | object | `{}` | Labels for controller pods |
| replicaCount | int | `1` | Number of replicas for the controller |
| resources | object | `{}` | Resource requests and limits for the controller |
| tolerations | list | `[]` | Tolerations for controller pods |
| serviceAccount.annotations | object | `{}` | Annotations for the controller service account |
| serviceAccount.create | bool | `true` | Create a service account to manage MimirRules CR |
| serviceAccount.name | string | `nil` | The name of the ServiceAccount to use for the controller. If not set and create is true, a name is generated. |
| controllerClusterRole.create | bool | `true` | Create a ClusterRole bound to the serviceAccount to manage MimirRules CR by the controller. |
| controllerClusterRole.annotations | object | `{}` | Annotations for the ClusterRole. |
| controllerClusterRole.name | string | `nil` | The name of ClusterRole bound to the serviceAccount. If not set and create is true, a name is generated. |
| crdClusterRoles.create | bool | `false` | Create a ClusterRole to manage MimirRules CR by users. |
| crdClusterRoles.annotations | object | `{}` | Annotations for the ClusterRole. |
| crdClusterRoles.name | string | `nil` | The name of ClusterRole. If not set and create is true, a name is generated. |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}}` | The SecurityContext for container |
| podSecurityContext | object | `{"runAsNonRoot":true}` | The SecurityContext for pod |
| service.annotations | object | `{}` | Annotations for the service |
| service.appProtocol | string | `nil` | Set appProtocol for the service |
| service.clusterIP | string | `nil` | ClusterIP of the service |
| service.labels | object | `{}` | Labels for service |
| service.loadBalancerIP | string | `nil` | Load balancer IPO address if service type is LoadBalancer |
| service.loadBalancerSourceRanges | list | `[]` | Load balancer allow traffic from CIDR list if service type is LoadBalancer |
| service.nodePort | string | `nil` | Node port if service type is NodePort |
| service.port | int | `80` | Port of the service |
| service.type | string | `"ClusterIP"` | Type of the service |
| serviceMonitor.annotations | object | `{}` | ServiceMonitor annotations |
| serviceMonitor.enabled | bool | `false` | If enabled, ServiceMonitor resources for Prometheus Operator are created |
| serviceMonitor.interval | string | `nil` | ServiceMonitor scrape interval |
| serviceMonitor.labels | object | `{}` | Additional ServiceMonitor labels |
| serviceMonitor.matchExpressions | list | `[]` | Optional expressions to match on |
| serviceMonitor.metricRelabelings | list | `[]` | ServiceMonitor metric relabel configs to apply to samples before ingestion https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#endpoint |
| serviceMonitor.namespace | string | `nil` | Alternative namespace for ServiceMonitor resources |
| serviceMonitor.namespaceSelector | object | `{}` | Namespace selector for ServiceMonitor resources |
| serviceMonitor.relabelings | list | `[]` | ServiceMonitor relabel configs to apply to samples before scraping https://github.com/prometheus-operator/prometheus-operator/blob/master/Documentation/api.md#relabelconfig |
| serviceMonitor.scheme | string | `"http"` | ServiceMonitor will use http by default, but you can pick https as well |
| serviceMonitor.scrapeTimeout | string | `nil` | ServiceMonitor scrape timeout in Go duration format (e.g. 15s) |
| serviceMonitor.targetLabels | list | `[]` | ServiceMonitor will add labels from the service to the Prometheus metric https://github.com/prometheus-operator/prometheus-operator/blob/main/Documentation/api.md#servicemonitorspec |
| serviceMonitor.tlsConfig | string | `nil` | ServiceMonitor will use these tlsConfig settings to make the health check requests |


Auth proxy settings not covered here.