# Helm repository for Mimir Rules Kubernetes controller

K8S controller to manage Mimir Alerting and Recording rules with dynamic tenants, based on namespace name or annotations. The controller uses an [CRD `MimirRules`](https://github.com/springernature/o11y-rules-telemetry-operator/blob/main/config/crd/bases/mimirrules.telemetry.springernature.com_mimirrules.yaml) which matches the [`PrometheusRule` specs](https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.PrometheusRule) defined by the [Prometheus Operator](https://prometheus-operator.dev/).

The current MimirRules specs are defined in the [repository, in `/config/crd/bases/` directory](https://github.com/springernature/o11y-rules-telemetry-operator/blob/main/config/crd/bases/mimirrules.telemetry.springernature.com_mimirrules.yaml).

> Warning!
>
> Installing this Helm Chart will install a new CRD `MimirRules` in the cluster. Take this into account for CRD updates or when deleting the controller, those operations will require manual steps with `kubectl delete/apply`.

## Installing

[Helm](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

Once Helm is set up properly, you have two ways to install the chart.

## 1. Using GitHub registry with Helm

To show the chart information and see the templates:

```console
helm show all oci://ghcr.io/springernature/charts/mimirrules-controller --version 1.0.3
helm template mimirrules-controller oci://ghcr.io/springernature/charts/mimirrules-controller --version 1.0.3
```

```console
helm upgrade --install mimirrules-controller oci://ghcr.io/springernature/charts/mimirrules-controller --version 1.0.3
```

## 2. Adding Helm repository

First, add this Helm repository:

```console
helm repo add o11y-rules-telemetry-operator https://springernature.github.io/o11y-rules-telemetry-operator
helm repo update
```

You can then run `helm search repo o11y-rules-telemetry-operator` to see the charts.

To install the app chart in the Kubernetes cluster.

```console
helm upgrade --install mimirrules-controller o11y-rules-telemetry-operator/mimirrules-controller --set config.mimirAPI=http://your.mimir.api
```

You can see all configuration parameters and default values of the Chart in the [Readme](https://github.com/springernature/o11y-rules-telemetry-operator/blob/main/charts/Readme.md)
