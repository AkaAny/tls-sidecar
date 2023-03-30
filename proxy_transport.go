package tls_sidecar

import "net/http"

type RoundTripFunc func(base http.RoundTripper, request *http.Request) (*http.Response, error)

type ProxyRoundTripper struct {
	base          http.RoundTripper
	roundTripFunc RoundTripFunc
}

func NewProxyRoundTripper(base http.RoundTripper, fn RoundTripFunc) http.RoundTripper {
	return &ProxyRoundTripper{
		base:          base,
		roundTripFunc: fn,
	}
}

func (b *ProxyRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	return b.roundTripFunc(b.base, request)
}
