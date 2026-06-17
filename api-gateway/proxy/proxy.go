package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// New cria um ReverseProxy que encaminha ao targetURL preservando path e headers.
func New(targetURL string) (*httputil.ReverseProxy, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("proxy: URL inválida %q: %w", targetURL, err)
	}

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host
			req.Header.Del("X-Forwarded-Host")
		},
		ModifyResponse: func(res *http.Response) error {
			return nil
		},
	}, nil
}
