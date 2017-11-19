#!/bin/bash
set -ex

PROJECT_DIR="$(cd "$(dirname "$0")/.."; pwd)"

go build -o ${PROJECT_DIR}/bin/apps-metrics-plugin ${PROJECT_DIR}/cmd/plugin/apps_metrics.go
cf uninstall-plugin AppsMetricsPlugin || true
cf install-plugin ${PROJECT_DIR}/bin/apps-metrics-plugin -f
