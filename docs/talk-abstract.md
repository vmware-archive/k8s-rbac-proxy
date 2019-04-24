RBAC Proxy: Non-intrusive Isolation of Controllers

---

Kubernetes controllers that need to work across multiple namespaces require querying of resources in those namespaces. This is done either by elevating permissions of the controller to cluster level or by looping through all the configured namespaces. First solution is not ideal from a security perspective and the second one requires the controller to offer namespace aware configuration which is often not available.

Ideally deployed controllers should remain oblivious to namespace requirements and be limited by their RBAC rules. One solution is to place a layer of abstraction between the API server and the controllers. This proxy server can emulate cluster level calls by returning resources only from namespaces accessible to the controller's service account. This layer of indirection imposes some performance overhead, but is less intrusive to the controller and easier to configure.

----

Running clusters used by multiple teams can become a challenge if they need to use same type of controllers (for example, both team A and team B need to use etcd-operator). One common way to address this problem is to create a lot of small clusters. An alternative way is to try to isolate controllers via RBAC to specific set of namespaces. By doing so common use cases become more manageable in a shared cluster:

- rolling out new versions of controllers so that their impact is isolated to specific set of resources (such as staging vs production resources, potentially spread around multiple namespaces)

- allowing teams to manage controllers in their own way without affecting other teams (for example having team A use etcd-operator v1 for all their namespaces, and team B use etcd-operator v2 for theirs)

Even though our current solution does not resolve conflicts in cluster level resources such as CRDs, it does make running a single larger shared cluster more accessible.

In this talk we will share:

- our experience using RBAC proxy with knative and etcd-operator
- go into its implementation details highlighting Kubernetes API decisions
- talk about security implications of using a proxy (proxy does not itself do any RBAC policy enforcement since it uses caller's service account tokens)
- showcase a live demo of RBAC proxy in action
- and, finally provide suggestions on how to manage shared clusters. 

We expect to open source our implementation of RBAC Proxy before the talk.
