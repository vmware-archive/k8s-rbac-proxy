package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListReq struct {
	originalPath string
	newPaths     []string
	pathMeta     map[string]string
	req          *http.Request
	respWriter   http.ResponseWriter

	chunksCh          chan ListChunk
	typeMetaResolver  TypeMetaResolver
	upstreamTransport *http.Transport

	logger Logger
}

type ListChunk struct {
	Error      error
	Header     http.Header
	StatusCode int
	Data       []byte
}

type ListEnvelope struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// plain response application/json
	Items []json.RawMessage `json:"items,omitempty"`

	// table response application/json;as=Table;v=v1beta1;g=meta.k8s.io
	ColDefs []json.RawMessage `json:"columnDefinitions,omitempty"`
	Rows    []json.RawMessage `json:"rows,omitempty"`
}

func (p ListReq) Respond() {
	if len(p.newPaths) == 0 {
		p.writeEmptyResp()
		return
	}

	p.logger.Debug("headers: %#v", p.req.Header)

	for _, path := range p.newPaths {
		go p.readSingle(path)
	}

	err := p.mergeChunks()
	if err != nil && err != io.EOF {
		p.logger.Error("fail: %s", err)
	} else {
		p.logger.Info("success")
	}
}

func (p ListReq) readSingle(path string) {
	newReq := copyReq(p.req)
	newReq.URL.Path = path

	p.logger.Info("sub request (request %s %s)", newReq.Method, newReq.URL.Path)

	resp, err := p.upstreamTransport.RoundTrip(newReq)
	if err != nil {
		p.chunksCh <- ListChunk{Error: err}
		return
	}

	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)

	p.logger.Debug("body (request %s %s): %s", newReq.Method, newReq.URL.Path, bs)

	p.chunksCh <- ListChunk{Error: err, Header: resp.Header, StatusCode: resp.StatusCode, Data: bs}
}

func (p ListReq) mergeChunks() error {
	var selectedHeaderChunk *ListChunk
	var envelopes []ListEnvelope

	for i := 0; i < cap(p.chunksCh); i++ {
		chunk := <-p.chunksCh

		if selectedHeaderChunk == nil || chunk.Error != nil {
			selectedHeaderChunk = &chunk
		}

		var envelope ListEnvelope
		err := json.Unmarshal(chunk.Data, &envelope)
		if err != nil {
			return fmt.Errorf("unmarshaling response envelope: %s", err)
		}

		envelopes = append(envelopes, envelope)
	}

	envelopes[0].SelfLink = p.originalPath
	// TODO limit, continue, revisionVersion

	for i, env := range envelopes {
		if i != 0 {
			envelopes[0].Items = append(envelopes[0].Items, env.Items...)
			// col defs just come from the first envelope
			envelopes[0].Rows = append(envelopes[0].Rows, env.Rows...)
		}
	}

	return p.writeResp(envelopes[0], *selectedHeaderChunk)
}

func (p ListReq) writeEmptyResp() {
	typeMeta, err := p.typeMetaResolver.Resolve(p.pathMeta)
	if err != nil {
		p.logger.Error("resolving meta: %s", err)
		p.respWriter.WriteHeader(http.StatusInternalServerError)
		p.respWriter.Write([]byte(""))
		return
	}

	envelope := ListEnvelope{
		TypeMeta: typeMeta,
		ListMeta: metav1.ListMeta{
			SelfLink:        p.originalPath,
			ResourceVersion: "1",
		},
	}

	headerChunk := ListChunk{
		Header: http.Header{
			http.CanonicalHeaderKey("content-type"): []string{"application/json"},
		},
		StatusCode: 200,
	}

	err = p.writeResp(envelope, headerChunk)
	if err != nil {
		p.logger.Error("fail: %s", err)
	} else {
		p.logger.Info("success")
	}
}

func (p ListReq) writeResp(envelope ListEnvelope, headerChunk ListChunk) error {
	newBs, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshaling response envelope: %s", err)
	}

	p.logger.Debug("final body: %s", newBs)

	copyHeaders(p.respWriter.Header(), headerChunk.Header)
	p.respWriter.Header()[http.CanonicalHeaderKey("Content-Length")] = []string{fmt.Sprintf("%d", len(newBs))} // TODO ugly
	p.respWriter.WriteHeader(headerChunk.StatusCode)                                                           // TODO err?

	_, err = p.respWriter.Write(newBs)
	if err != nil {
		return fmt.Errorf("writing response: %s", err)
	}

	return nil
}
