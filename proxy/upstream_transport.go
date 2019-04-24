package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
)

type UpstreamTransportFactory struct {
	transport *http.Transport
}

func NewUpstreamTransportFactory(logger Logger) (UpstreamTransportFactory, error) {
	const kubeCAPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	certs, err := ioutil.ReadFile(kubeCAPath)
	if err != nil {
		return UpstreamTransportFactory{}, fmt.Errorf("Appending %q to RootCAs: %v", kubeCAPath, err)
	}

	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		return UpstreamTransportFactory{}, fmt.Errorf("Appending root CA certs")
	}

	config := &tls.Config{
		RootCAs: rootCAs,
	}

	return UpstreamTransportFactory{&http.Transport{TLSClientConfig: config}}, nil
}

func (f UpstreamTransportFactory) New() *http.Transport { return f.transport }
