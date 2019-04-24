#!/bin/bash

set -e -o pipefail

cd "$(dirname "$0")/.."

apt-get -y update
apt-get -y install wget perl

./hack/install-kubectl-in-cluster.sh
./hack/apply-config-in-cluster.sh
