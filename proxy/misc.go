package proxy

import (
	"net/http"
)

func copyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func copyReq(req *http.Request) *http.Request {
	newURL := *req.URL
	return &http.Request{
		Method:           req.Method,           // string
		URL:              &newURL,              // *url.URL
		Proto:            req.Proto,            // string
		ProtoMajor:       req.ProtoMajor,       // int
		ProtoMinor:       req.ProtoMinor,       // int
		Header:           req.Header,           // Header
		ContentLength:    req.ContentLength,    // int64
		TransferEncoding: req.TransferEncoding, // []string
		Close:            req.Close,            // bool
		Host:             req.Host,             // string
	}
}
