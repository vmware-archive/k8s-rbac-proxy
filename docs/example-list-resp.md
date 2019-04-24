## Example list request/resp

```
GET /api/v1/namespaces/rbac-test/pods
GET /apis/rbac.authorization.k8s.io/v1/namespaces/rbac-test/rolebindings
```

```json
{
  "kind": "PodList",
  "apiVersion": "v1",
  "metadata": {
    "selfLink": "/api/v1/namespaces/rbac-test/pods",
    "resourceVersion": "15575434"
  },
  "items": [
    {
      "metadata": {
        "name": "w-8f4r9",
        "generateName": "w-",
        "namespace": "rbac-test",
        "selfLink": "/api/v1/namespaces/rbac-test/pods/w-8f4r9",
        "uid": "47c8d6d3-1533-11e9-b33a-42010a800090",
        "resourceVersion": "15575375",
        "creationTimestamp": "2019-01-10T23:56:00Z",
        "labels": {
          "foo2": "",
          "foo3": "",
          "kwt.cppforlife.com/workspace": "true"
        },
        "annotations": {
          "kwt.cppforlife.com/workspace-last-used": "2019-01-11T18:37:15Z"
        }
      },
      "spec": {
        "volumes": [
          {
            "name": "rbac-pods-token-ftmxp",
            "secret": {
              "secretName": "rbac-pods-token-ftmxp",
              "defaultMode": 420
            }
          }
        ],
        "containers": [
          {
            "name": "debug",
            "image": "ubuntu:xenial",
            "command": [
              "/bin/bash"
            ],
            "args": [
              "-c",
              "while true; do sleep 86400; done"
            ],
            "workingDir": "/tmp/build/ary23",
            "resources": {},
            "volumeMounts": [
              {
                "name": "rbac-pods-token-ftmxp",
                "readOnly": true,
                "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount"
              }
            ],
            "terminationMessagePath": "/dev/termination-log",
            "terminationMessagePolicy": "File",
            "imagePullPolicy": "IfNotPresent"
          }
        ],
        "restartPolicy": "Never",
        "terminationGracePeriodSeconds": 30,
        "dnsPolicy": "ClusterFirst",
        "serviceAccountName": "rbac-pods",
        "serviceAccount": "rbac-pods",
        "nodeName": "gke-dkalinin-oct-26-default-pool-18fe4288-z28d",
        "securityContext": {},
        "schedulerName": "default-scheduler",
        "tolerations": [
          {
            "key": "node.kubernetes.io/not-ready",
            "operator": "Exists",
            "effect": "NoExecute",
            "tolerationSeconds": 300
          },
          {
            "key": "node.kubernetes.io/unreachable",
            "operator": "Exists",
            "effect": "NoExecute",
            "tolerationSeconds": 300
          }
        ]
      },
      "status": {
        "phase": "Running",
        "conditions": [
          {
            "type": "Initialized",
            "status": "True",
            "lastProbeTime": null,
            "lastTransitionTime": "2019-01-10T23:56:00Z"
          },
          {
            "type": "Ready",
            "status": "True",
            "lastProbeTime": null,
            "lastTransitionTime": "2019-01-10T23:56:02Z"
          },
          {
            "type": "PodScheduled",
            "status": "True",
            "lastProbeTime": null,
            "lastTransitionTime": "2019-01-10T23:56:00Z"
          }
        ],
        "hostIP": "10.128.0.10",
        "podIP": "10.20.4.178",
        "startTime": "2019-01-10T23:56:00Z",
        "containerStatuses": [
          {
            "name": "debug",
            "state": {
              "running": {
                "startedAt": "2019-01-10T23:56:01Z"
              }
            },
            "lastState": {},
            "ready": true,
            "restartCount": 0,
            "image": "ubuntu:xenial",
            "imageID": "docker-pullable://ubuntu@sha256:e547ecaba7d078800c358082088e6cc710c3affd1b975601792ec701c80cdd39",
            "containerID": "docker://1d75ad4874dc75719d96da2520b0fc56772d3a8bd79c9e7c7b0c1b57c182f339"
          }
        ],
        "qosClass": "BestEffort"
      }
    }
  ]
}
```
