#!/usr/bin/env bash
set -ex

GOOS=linux GOARCH=amd64 go build .
cf push
rm prometheus