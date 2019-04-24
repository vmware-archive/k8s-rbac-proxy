#!/bin/bash

set -e -o pipefail

export GOPATH=$PWD

cd "$(dirname "$0")/.."

export PATH=/usr/local/go/bin:$PATH
export CGO_ENABLED=0 # no gcc on path
export KUBERNETES_SERVICE_HOST=10.23.240.169 # rbac-proxy.rbac-proxy.svc.cluster.local
# export KUBERNETES_SERVICE_PORT=444

GOCACHE=off go test ./test/e2e/ -timeout 120m -test.v

echo "Success"
