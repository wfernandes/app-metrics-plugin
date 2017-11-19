#!/bin/bash
set -e

PROJECT_DIR="$(cd "$(dirname "$0")/.."; pwd)"

pushd ${PROJECT_DIR}/cmd
    ginkgo -r -race
popd

pushd ${PROJECT_DIR}/pkg
    ginkgo -r -race
popd