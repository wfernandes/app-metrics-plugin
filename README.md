[![CI Badge][ci-badge]][ci-badge-link]
# App Metrics Plugin

This is a CF CLI plugin. See [here][cf-cli] for more details.

This plugin hits your app's metrics endpoint **across all your app instances** and displays human readable output.

## Purpose
Instrumenting applications with various metrics to provide insight into its operation is an important practice for
modern day developers. I shouldn't have to explain this further.


The purpose of this plugin is to be able hit the metrics endpoint across all your application instances.
There are many ways to instrument your applications including [Dropwizard][dropwizard], [Prometheus][prometheus] and
[go-metrics][godropwizard].


For golang applications, there is a package called [Expvar][expvar] available as part of the standard library which
makes your metrics available via a HTTP endpoint

### Problem this plugin alleviates

So if you have an app deployed on Cloudfoundry at `myapp.domain.cf-app.com` you can `GET` its metrics by hitting
`http://myapp.domain.cf-app.com/metrics`.

However, if you scale up the application to greater than one instance, Cloudfoundry's router will load balance your requests across
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
   app-metrics - Hits the metrics endpoint across all your app instances

USAGE:
   cf app-metrics APP_NAME

OPTIONS:
   -endpoint       path of the metrics endpoint
   -template       path of the template files to render metrics

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
$ cf app-metrics event-alerts

Instance: 0
Metrics:
  eva.notificationsStatus: 1
  eva.smoketest: 5
  ingress.matched: 0
  ingress.received: 1131578
  notifier.dropped: 0
  notifier.emails.failed: 0
  notifier.emails.sent: 0

Instance: 1
Metrics:
  eva.notificationsStatus: 1
  eva.smoketest: 0
  ingress.matched: 0
  ingress.received: 24521
  notifier.dropped: 0
  notifier.emails.failed: 0
  notifier.emails.sent: 0


```

## Future Work

See the issues section for thoughts on what needs to be added later.


## Releasing

In order to create a new release, follow these steps

1. Create local tag and binaries
  ```
  ./scripts/build-all.sh release VERSION_NUMBER #(e.g. 0.7.0)
  ```
1. Copy the output of the previous command from the first line (should be '- name: AppsMetrics Plugin' to the last checksum line (should be something like checksum: fde5fd52c40ea4c34330426c09c143a76a77a8db)
1. Push the tag `git push --follow-tags`
1. On github, create new release based on new tag [here](https://github.com/cloudfoundry-community/firehose-plugin/releases/new)
1. Upload the three binaries from the ./bin folders to the release (Linux, OSX and Win64)
1. Fork [this repo](https://github.com/cloudfoundry-incubator/cli-plugin-repo) and clone it locally
1. Edit the repo-index.yml
  ```
  vi repo-index.yml
  ```
  to override the existing section about the firehose plugin with the text previously copied in Step 2.
1. Push the change to your fork
1. Create a PR against the [original repo](https://github.com/cloudfoundry-incubator/cli-plugin-repo/compare)


[ci-badge-link]:    https://travis-ci.org/wfernandes/app-metrics-plugin
[ci-badge]:         https://travis-ci.org/wfernandes/app-metrics-plugin.svg?branch=master
[cf-cli]:           https://docs.cloudfoundry.org/cf-cli/develop-cli-plugins.html
[dropwizard]:       http://metrics.dropwizard.io/3.2.3/
[prometheus]:       https://prometheus.io/docs/practices/instrumentation/
[expvar]:           https://golang.org/pkg/expvar/
[godropwizard]:     https://github.com/rcrowley/go-metrics
