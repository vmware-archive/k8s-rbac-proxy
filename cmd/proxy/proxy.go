package main

import (
	"flag"
	"os"

	"github.com/cppforlife/kube-rbac-proxy/proxy"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	listenAddr  = flag.String("listenAddr", ":443", "Listen address")
	tlsCertPath = flag.String("tlsCertPath", "", "Certificate path")
	tlsKeyPath  = flag.String("tlsKeyPath", "", "Private key path")
	debug       = flag.Bool("debug", false, "Print debug log")
	pprof       = flag.Bool("pprof", false, "Add pprof endpoints")
)

func main() {
	flag.Parse()

	logger := proxy.NewOutLogger(*debug)

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	// TODO high QPS
	config.QPS = 1000
	config.Burst = 1000

	coreClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	upstreamTransportFactory, err := proxy.NewUpstreamTransportFactory(logger)
	if err != nil {
		panic(err.Error())
	}

	pr := proxy.Proxy{
		ListenAddr:   *listenAddr,
		CertFilePath: *tlsCertPath,
		KeyFilePath:  *tlsKeyPath,

		Pprof: *pprof,

		UpstreamAPIServerHost:    os.Getenv("KUBERNETES_SERVICE_HOST"),
		ServiceAccountFactory:    proxy.NewServiceAccountFactory(coreClient),
		UpstreamTransportFactory: upstreamTransportFactory,
		TypeMetaResolver:         proxy.NewTypeMetaResolver(coreClient),

		Logger: logger,
	}

	logger.Info("Starting proxy...")

	err = pr.Run()
	if err != nil {
		panic(err.Error())
	}
}
