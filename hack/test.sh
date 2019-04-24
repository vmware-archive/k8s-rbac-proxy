#!/bin/bash

set -e

kubectl delete -f hack/rbac-proxy-tester.yml || true
sleep 5
kubectl apply -f hack/rbac-proxy-tester.yml

kwt w create \
	-n default --service-account rbac-proxy-tester \
	--install-go1x \
	-i kube-rbac-proxy=.:src/github.com/cppforlife/kube-rbac-proxy \
	--command src/github.com/cppforlife/kube-rbac-proxy/hack/test-in-cluster.sh \
	--rm

kubectl delete -f hack/rbac-proxy-tester.yml
