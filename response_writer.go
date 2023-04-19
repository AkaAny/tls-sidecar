package tls_sidecar

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// NewResponseWriter create a http response writer
func NewResponseWriter(protoMajor, protoMinor int) *ResponseWriter {
	return &ResponseWriter{protoMajor: protoMajor, protoMinor: protoMinor}
}

type ResponseWriter struct {
	protoMajor, protoMinor int
	statusCode             int
	header                 http.Header
	body                   bytes.Buffer
}

func (r *ResponseWriter) Header() http.Header {
	if nil == r.header {
		r.header = make(http.Header)
		r.header.Set("server", "go-netty")
	}
	return r.header
}

func (r *ResponseWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *ResponseWriter) Write(b []byte) (int, error) {
	return r.body.Write(b)
}

func (r *ResponseWriter) Response() *http.Response {
	if 0 == r.statusCode {
		r.WriteHeader(http.StatusOK)
	}
	return &http.Response{
		ProtoMajor:    r.protoMajor,
		ProtoMinor:    r.protoMinor,
		StatusCode:    r.statusCode,
		Header:        r.Header(),
		Body:          ioutil.NopCloser(&r.body),
		ContentLength: int64(r.body.Len()),
	}
}
