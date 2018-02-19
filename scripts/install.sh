#!/bin/bash
set -ex

PROJECT_DIR="$(cd "$(dirname "$0")/.."; pwd)"

go build -o ${PROJECT_DIR}/bin/app-metrics-plugin ${PROJECT_DIR}/cmd/plugin/app_metrics.go
cf uninstall-plugin app-metrics || true
cf install-plugin ${PROJECT_DIR}/bin/app-metrics-plugin -f
