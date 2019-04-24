package proxy

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"time"
)

type Proxy struct {
	ListenAddr                string
	CertFilePath, KeyFilePath string

	Pprof bool

	UpstreamAPIServerHost    string // os.Getenv("KUBERNETES_SERVICE_HOST")
	ServiceAccountFactory    ServiceAccountFactory
	UpstreamTransportFactory UpstreamTransportFactory
	TypeMetaResolver         TypeMetaResolver

	Logger Logger
}

func (p Proxy) handler(w http.ResponseWriter, req *http.Request) {
	reqID := fmt.Sprintf("req-%d", time.Now().UTC().UnixNano())

	req.URL.Scheme = "https"
	req.URL.Host = p.UpstreamAPIServerHost

	p.Logger.Info("[%s] responding (request %s %s watch=%t)", reqID, req.Method, req.URL.Path, p.isWatch(req))

	upstreamTransport := p.UpstreamTransportFactory.New()

	resp, err := upstreamTransport.RoundTrip(req)
	if err != nil {
		p.Logger.Error("[%s] initial roundtrip (request %s %s): %s", reqID, req.Method, req.URL.Path, err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(""))
		return
	}

	defer resp.Body.Close()

	if req.Method == "GET" && resp.StatusCode == http.StatusForbidden {
		if p.streamMergedResp(reqID, req, w) {
			return
		}
	}

	p.streamOpaqueResp(reqID, resp, req, w)
}

func (p Proxy) streamOpaqueResp(reqID string, resp *http.Response, req *http.Request, w http.ResponseWriter) {
	copyHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)

	var writer io.Writer

	// Only flush on watching request as it's wasteful otherwise
	if p.isWatch(req) {
		writer = flushingWriter{w}
	} else {
		writer = w
	}

	_, err := io.Copy(writer, resp.Body)
	if err != nil {
		p.Logger.Error("[%s] stream opaque resp (request %s %s): %s", reqID, req.Method, req.URL.Path, err)
		return
	}

	p.Logger.Info("[%s] stream opaque success (request %s %s)", reqID, req.Method, req.URL.Path)
}

func (p Proxy) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", p.handler)

	if p.Pprof {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	cfg := &tls.Config{}
	srv := &http.Server{
		Addr:         p.ListenAddr,
		Handler:      mux,
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	return srv.ListenAndServeTLS(p.CertFilePath, p.KeyFilePath)
}

type flushingWriter struct {
	w http.ResponseWriter
}

func (l flushingWriter) Write(data []byte) (int, error) {
	n, err := l.w.Write(data)
	l.w.(http.Flusher).Flush()
	return n, err
}
