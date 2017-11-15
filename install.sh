#!/bin/bash
set -ex

go build -o apps-metrics-plugin main.go
cf uninstall-plugin AppsMetricsPlugin
cf install-plugin apps-metrics-plugin -f
