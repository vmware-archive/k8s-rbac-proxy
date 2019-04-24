## etcd-operator example

- 5 namespaces
  - etcd-operator1 -> has etcd-operator deployment and other admin things
    - teama-etcds -> contains etcd clusters for team a
    - teamb-etcds -> contains more etcd clusters for team b
  - etcd-operator2 -> has etcd-operator deployment and other admin things
    - teamc-etcds -> contains etcd clusters for team a

Configs are based on https://github.com/coreos/etcd-operator/blob/master/example/ (etcd configuration is not tuned in any way!!!).

```
kapp deploy -a etcd-operator1 -f examples/etcd-operator/100-etcd-operator1.yml \
  --allow-cluster --allow-ns etcd-operator1

kapp deploy -a teama-etcds -f examples/etcd-operator/200-teama-etcds.yml \
  --allow-ns teama-etcds --allow-cluster

kapp deploy -a teamb-etcds -f examples/etcd-operator/300-teamb-etcds.yml --allow-ns teamb-etcds --allow-cluster

kapp deploy -a etcd-operator2 -f examples/etcd-operator/400-etcd-operator2.yml \
  --allow-cluster --allow-ns etcd-operator2

kapp deploy -a teamc-etcds -f examples/etcd-operator/500-teamc-etcds.yml \
  --allow-ns teamc-etcds --allow-cluster
```
