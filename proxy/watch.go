package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type WatchRequest struct {
	newPaths   []string
	req        *http.Request
	respWriter http.ResponseWriter

	chunksCh          chan WatchEventChunk
	cancelCh          chan struct{}
	chunksWg          sync.WaitGroup
	upstreamTransport *http.Transport

	logger Logger
}

type WatchEventChunk struct {
	Error      error
	Header     http.Header
	StatusCode int
	Data       []byte
}

func (p *WatchRequest) Stream() {
	for _, path := range p.newPaths {
		path := path
		p.chunksWg.Add(1)

		go func() {
			defer p.chunksWg.Done()
			p.readSingleStream(path)
		}()
	}

	if len(p.newPaths) == 0 {
		go p.writeEmptyChunk()
	}

	err := p.writeChunks()
	if err != nil {
		if err != io.EOF {
			p.logger.Error("fail: %s", err)
		} else {
			p.logger.Info("success")
		}
	}

	go p.cancelStreams()
}

func (p *WatchRequest) Cancel() {
	p.logger.Info("canceling")

	p.chunksCh <- WatchEventChunk{
		Header: http.Header{
			// TODO do we need to return audit id?
			http.CanonicalHeaderKey("content-type"): []string{"application/json"},
		},
		StatusCode: 200,
		Error:      fmt.Errorf("watched namespaces changed"),
	}
}

func (p *WatchRequest) readSingleStream(path string) {
	newReq := copyReq(p.req)
	newReq.URL.Path = path

	p.logger.Info("sub request (request %s %s)", newReq.Method, newReq.URL.Path)

	resp, err := p.upstreamTransport.RoundTrip(newReq)
	if err != nil {
		select {
		case <-p.cancelCh:
			// chunksCh was closed
		default:
			p.chunksCh <- WatchEventChunk{Error: err}
		}
		return
	}

	defer resp.Body.Close()
	chunkReader := bufio.NewReader(resp.Body)

	for {
		bs, err := chunkReader.ReadBytes('\n')

		p.logger.Debug("event body (request %s %s): %s", newReq.Method, newReq.URL.Path, bs)

		select {
		case <-p.cancelCh:
			return // exit quickly so that watch stream is closed
		default:
			p.chunksCh <- WatchEventChunk{Error: err, Header: resp.Header, StatusCode: resp.StatusCode, Data: bs}
			if err != nil {
				return
			}
		}
	}
}

func (p *WatchRequest) writeEmptyChunk() {
	p.chunksCh <- WatchEventChunk{
		Header: http.Header{
			// TODO do we need to return audit id?
			http.CanonicalHeaderKey("content-type"): []string{"application/json"},
		},
		StatusCode: 200,
	}
}

func (p *WatchRequest) writeChunks() error {
	wantsHeader := true
	for chunk := range p.chunksCh {
		if wantsHeader {
			copyHeaders(p.respWriter.Header(), chunk.Header)
			p.respWriter.WriteHeader(chunk.StatusCode)
			wantsHeader = false
		}
		if chunk.Error != nil {
			return chunk.Error
		}
		_, err := p.respWriter.Write(chunk.Data)
		p.respWriter.(http.Flusher).Flush()
		if err != nil {
			return fmt.Errorf("writing chunk data: %s", err)
		}
	}
	return nil
}

func (p *WatchRequest) cancelStreams() {
	// Cancel at next possible time, possibly still open watch requests
	close(p.cancelCh)
	for {
		<-p.chunksCh
	}
	p.chunksWg.Wait()
	close(p.chunksCh)
	p.logger.Info("complete")
}

type WatchChanges struct {
	initialNssByName map[string]struct{}
	serviceAccount   ServiceAccount
	watchReq         *WatchRequest
	stopCh           chan struct{}
	logger           Logger
}

func NewWatchChanges(nss []string, serviceAccount ServiceAccount, watchReq *WatchRequest, logger Logger) WatchChanges {
	initialNssByName := map[string]struct{}{}
	for _, name := range nss {
		initialNssByName[name] = struct{}{}
	}
	return WatchChanges{initialNssByName, serviceAccount, watchReq, make(chan struct{}), logger}
}

func (c WatchChanges) Observe() {
	for {
		time.Sleep(10 * time.Second)

		select {
		case <-c.stopCh:
			return
		default:
			// continue checking
		}

		newNss, err := c.serviceAccount.Namespaces()
		if err != nil || c.nssDifferent(newNss) {
			c.watchReq.Cancel()
			return
		}
	}
}

func (c WatchChanges) nssDifferent(newNss []string) bool {
	if len(c.initialNssByName) != len(newNss) {
		c.logger.Debug("canceling: nss len changed: %#v -> %v", c.initialNssByName, newNss)
		return true
	}
	for _, name := range newNss {
		if _, found := c.initialNssByName[name]; !found {
			c.logger.Debug("canceling: ns %s not found", name)
			return true
		}
	}
	return false
}

func (c WatchChanges) StopObserving() {
	close(c.stopCh)
}
