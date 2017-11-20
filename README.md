# Apps Metrics Plugin

This is a CF CLI plugin. See [here][cf-cli] for more details.

## Purpose
Instrumenting applications with various metrics to provide insight into its operation is an important practice for
modern day developers. I shouldn't have to explain this further.


The purpose of this plugin is to be able hit the metrics endpoint across all your application instances.
There are many ways to instrument your applications including [Dropwizard][dropwizard] and [Prometheus][prometheus].


For golang applications, there is a package called [Expvar][expvar] available as part of the standard library which
makes your metrics available via a HTTP endpoint

### Expvar

This plugin **currently** parses out expvar style metrics endpoint only. By default it hides the properties `cmdline`
and `memstats` as they tend to clutter up the output.

## Install
```bash
# To install a dev build of the plugin
./scripts/install.sh
```

## Usage

```
NAME:
   apps-metrics - Hits the metrics endpoint across all your app instances

USAGE:
   cf apps-metrics APP_NAME

OPTIONS:
   -endpoint       path of the metrics endpoint

```

## Uninstall

```bash
cf uninstall-plugin AppsMetricsPlugin
```

## Tests

```bash
./scripts/test.sh
```

## Future Work

See the issues section for thoughts on what needs to be added later.


[cf-cli]:       https://docs.cloudfoundry.org/cf-cli/develop-cli-plugins.html
[dropwizard]:   http://metrics.dropwizard.io/3.2.3/
[prometheus]:   https://prometheus.io/docs/practices/instrumentation/
[expvar]:       https://golang.org/pkg/expvar/