package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// New cria um ReverseProxy que encaminha ao targetURL preservando path e headers.
func New(targetURL string) *httputil.ReverseProxy {
	target, err := url.Parse(targetURL)
	if err != nil {
		panic("proxy: URL inválida: " + targetURL)
	}

	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.Host = target.Host
			// Remove header que vazaria endereço interno ao serviço downstream.
			req.Header.Del("X-Forwarded-Host")
		},
		ModifyResponse: func(res *http.Response) error {
			return nil
		},
	}
}
