#!/bin/bash

set -e -o pipefail

# Needed for interpolation
wget -O- https://github.com/cloudfoundry/bosh-cli/releases/download/v5.4.0/bosh-cli-5.4.0-linux-amd64 > /usr/local/bin/bosh
echo "ecc1b6464adf9a0ede464b8699525a473e05e7205357e4eb198599edf1064f57  /usr/local/bin/bosh" | shasum -a 256 -c -

wget -O- https://pkg.cfssl.org/R1.2/cfssl_linux-amd64 > /usr/local/bin/cfssl
echo "eb34ab2179e0b67c29fd55f52422a94fe751527b06a403a79325fed7cf0145bd  /usr/local/bin/cfssl" | shasum -a 256 -c -

wget -O- https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64 > /usr/local/bin/cfssljson
echo "1c9e628c3b86c3f2f8af56415d474c9ed4c8f9246630bd21c3418dbe5bf6401e  /usr/local/bin/cfssljson" | shasum -a 256 -c -

chmod +x /usr/local/bin/bosh /usr/local/bin/cfssl*

kubectl apply -f ./config/100-ns.yml

cat <<EOF | cfssl genkey - | cfssljson -bare server
{
  "hosts": [
    "rbac-proxy.rbac-proxy.svc.cluster.local"
  ],
  "CN": "rbac-proxy.rbac-proxy.svc.cluster.local",
  "key": {
    "algo": "ecdsa",
    "size": 256
  }
}
EOF

# Public key isnt regenerated if csr is updated
kubectl delete csr/rbac-proxy || true

bosh int ./config/200-csr.yml --var "tls_csr=$(cat server.csr|base64|tr -d '\n')" | kubectl apply -f -

kubectl certificate approve rbac-proxy

sleep 10

bosh int ./config/300-secret.yml \
	--var-file tls_cert=<(kubectl get csr/rbac-proxy -o yaml|bosh int - --path /status/certificate|base64 -d) \
	--var-file tls_key=server-key.pem | kubectl apply -f -

kubectl apply -f ./config/400-deployment.yml
