# Apps Metrics Plugin

This is a CF CLI plugin. See [here][cf-cli] for more details.

## Purpose
Instrumenting applications with various metrics to provide insight into its operation is an important practice for
modern day developers. I shouldn't have to explain this further.


The purpose of this plugin is to be able hit the metrics endpoint across all your application instances.
There are many ways to instrument your applications including [Dropwizard][dropwizard], [Prometheus][prometheus] and
[go-metrics][godropwizard].


For golang applications, there is a package called [Expvar][expvar] available as part of the standard library which
makes your metrics available via a HTTP endpoint

### Problem this plugin alleviates

So if you have an app deployed on Cloudfoundry at `myapp.domain.cf-app.com` you can `GET` its metrics endpoint hitting
`http://myapp.domain.cf-app.com/metrics`.

However, if you scale the app to greater than one instance, Cloudfoundry's router will load balance your requests across
your app instances. This means that you will be getting metrics from a random instance every time.


This plugin will hit your metrics endpoint (defaults to `/debug/metrics`) across all your app instances and display the
output in a human readable fashion.

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

## Sample Output
```bash
# Yes, I know. Templating needs some more work.
$ cf apps-metrics event-alerts

        Instance: 0
        Metrics:
          ingress.matched: 0
          ingress.received: 944066
          notifier.dropped: 0
          notifier.emails.failed: 0
          notifier.emails.sent: 0

        Instance: 1
        Metrics:
          ingress.matched: 0
          ingress.received: 301891
          notifier.dropped: 0
          notifier.emails.failed: 0
          notifier.emails.sent: 0

```

## Future Work

See the issues section for thoughts on what needs to be added later.


[cf-cli]:       https://docs.cloudfoundry.org/cf-cli/develop-cli-plugins.html
[dropwizard]:   http://metrics.dropwizard.io/3.2.3/
[prometheus]:   https://prometheus.io/docs/practices/instrumentation/
[expvar]:       https://golang.org/pkg/expvar/
[godropwizard]: https://github.com/rcrowley/go-metrics