#!/bin/bash

set -e -o pipefail

ver=$(wget -q -O - https://storage.googleapis.com/kubernetes-release/release/stable.txt)
wget -O- https://storage.googleapis.com/kubernetes-release/release/${ver}/bin/linux/amd64/kubectl > /usr/local/bin/kubectl2
# TODO checksum

cat <<'EOF' >/usr/local/bin/kubectl
#!/bin/bash
exec /usr/local/bin/kubectl2 --token=`cat /var/run/secrets/kubernetes.io/serviceaccount/token` "$@"
EOF

chmod +x /usr/local/bin/kubectl*

kubectl config set-cluster local --server=https://$KUBERNETES_SERVICE_HOST \
	--certificate-authority=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
kubectl config set-context local --cluster=local
kubectl config use-context local

# Check that it works
kubectl get node
