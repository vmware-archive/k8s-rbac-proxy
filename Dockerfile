from golang:1.11-alpine3.8 as builder
workdir /go/src/github.com/cppforlife/kube-rbac-proxy
copy . .
run ./hack/build.sh

from alpine:3.8
copy --from=builder /go/src/github.com/cppforlife/kube-rbac-proxy/out/proxy /proxy
cmd ["/proxy"]
