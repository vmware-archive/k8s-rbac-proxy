#!/bin/sh

set -e

go fmt github.com/cppforlife/kube-rbac-proxy/...

export GOOS=linux GOARCH=amd64

go build -o out/proxy github.com/cppforlife/kube-rbac-proxy/cmd/proxy/
go build -o out/client github.com/cppforlife/kube-rbac-proxy/cmd/client/
