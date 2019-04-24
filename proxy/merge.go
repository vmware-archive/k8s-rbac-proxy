package proxy

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"
)

var (
	nativeTypes = regexp.MustCompile("\\A/api/(?P<version>[^/]+)/(?P<resource>[^/]+)\\z")
	customTypes = regexp.MustCompile("\\A/apis/(?P<group>[^/]+)/(?P<version>[^/]+)/(?P<resource>[^/]+)\\z")
)

func (p Proxy) streamMergedResp(reqID string, req *http.Request, w http.ResponseWriter) bool {
	nativeMatch := nativeTypes.MatchString(req.URL.Path)
	customMatch := customTypes.MatchString(req.URL.Path)

	if !nativeMatch && !customMatch {
		return false
	}

	serviceAccount, err := p.ServiceAccountFactory.New(req.Header[http.CanonicalHeaderKey("authorization")])
	if err != nil {
		p.Logger.Error("[%s] building service account (request %s %s): %s", reqID, req.Method, req.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(""))
		return true
	}

	nss, err := serviceAccount.Namespaces()
	if err != nil {
		p.Logger.Error("[%s] fetching namespaces (request %s %s): %s", reqID, req.Method, req.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(""))
		return true
	}

	var newPaths []string
	var pathMeta map[string]string

	if nativeMatch {
		pathMeta = p.findNamedMatches(nativeTypes, req.URL.Path)
	}
	if customMatch {
		pathMeta = p.findNamedMatches(customTypes, req.URL.Path)
	}

	for _, ns := range nss {
		if nativeMatch {
			newPaths = append(newPaths, fmt.Sprintf("/api/%s/namespaces/%s/%s",
				pathMeta["version"], ns, pathMeta["resource"]))
		}
		if customMatch {
			newPaths = append(newPaths, fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s",
				pathMeta["group"], pathMeta["version"], ns, pathMeta["resource"]))
		}
	}

	if p.isWatch(req) {
		p.streamWatchResp(newPaths, nss, serviceAccount, reqID, req, w)
	} else {
		p.streamListResp(req.URL.Path, newPaths, pathMeta, reqID, req, w)
	}

	return true
}

func (p Proxy) streamWatchResp(newPaths []string, nss []string, serviceAccount ServiceAccount, reqID string, req *http.Request, w http.ResponseWriter) {
	logger := NewPrefixLogger(fmt.Sprintf("[%s] watch req (request %s %s)", reqID, req.Method, req.URL.Path), p.Logger)

	watchReq := &WatchRequest{
		newPaths:   newPaths,
		req:        req,
		respWriter: w,

		chunksCh:          make(chan WatchEventChunk),
		cancelCh:          make(chan struct{}),
		chunksWg:          sync.WaitGroup{},
		upstreamTransport: p.UpstreamTransportFactory.New(),

		logger: logger,
	}

	changes := NewWatchChanges(nss, serviceAccount, watchReq, NewPrefixLogger("watch changes", logger))

	go changes.Observe()
	defer changes.StopObserving()

	watchReq.Stream()
}

func (p Proxy) streamListResp(originalPath string, newPaths []string, pathMeta map[string]string, reqID string, req *http.Request, w http.ResponseWriter) {
	listReq := ListReq{
		originalPath: originalPath,
		newPaths:     newPaths,
		pathMeta:     pathMeta,
		req:          req,
		respWriter:   w,

		chunksCh:          make(chan ListChunk, len(newPaths)),
		typeMetaResolver:  p.TypeMetaResolver,
		upstreamTransport: p.UpstreamTransportFactory.New(),

		logger: NewPrefixLogger(fmt.Sprintf("[%s] list req (request %s %s)", reqID, req.Method, req.URL.Path), p.Logger),
	}

	listReq.Respond()
}

func (Proxy) findNamedMatches(reg *regexp.Regexp, str string) map[string]string {
	match := reg.FindStringSubmatch(str)
	results := map[string]string{}
	for i, name := range match {
		results[reg.SubexpNames()[i]] = name
	}
	return results
}

func (p Proxy) isWatch(req *http.Request) bool {
	watchQuery := req.URL.Query()["watch"]
	return len(watchQuery) > 0 && watchQuery[0] == "true"
}
