#!/bin/bash

# do not set -e

kubectl apply -f hack/rbac-proxy-deployer.yml

kwt w c -n default -i kube-rbac-proxy=. \
	--command kube-rbac-proxy/hack/deploy-in-cluster.sh --rm \
	--service-account rbac-proxy-deployer

kubectl delete -f hack/rbac-proxy-deployer.yml
