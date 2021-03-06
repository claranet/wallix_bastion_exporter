package httpclient

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// The configuration to build the HTTP client.
type HTTPConfig struct {
	Timeout       int
	Username      string
	Password      string
	Headers       map[string]string
	SkipVerify    bool
	CookieManager bool
}

// An http transport that injects basic auth into each request.
type TransportWithBasicAuth struct {
	http.RoundTripper
	Username string
	Password string
}

// Override the only method that the client actually calls on the transport to
// do the request.
func (t *TransportWithBasicAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.Username, t.Password)

	return t.RoundTripper.RoundTrip(req)
}

// Build returns a configured http.Client.
func (h *HTTPConfig) Build() (client *http.Client, err error) {
	roundTripper := func() http.RoundTripper {
		transport := http.DefaultTransport.(*http.Transport).Clone()

		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: h.SkipVerify, //nolint:gosec
		}

		return transport
	}()

	if h.Username != "" {
		roundTripper = &TransportWithBasicAuth{
			RoundTripper: roundTripper,
			Username:     h.Username,
			Password:     h.Password,
		}
	}

	if h.Headers == nil {
		h.Headers = map[string]string{}
	}

	roundTripper = &addHeader{
		rt:      roundTripper,
		headers: h.Headers,
	}

	client = &http.Client{
		Timeout:   time.Second * time.Duration(h.Timeout),
		Transport: roundTripper,
	}

	// Handle cookie authentication
	if h.CookieManager {
		cookieJar, err := cookiejar.New(nil)
		if err != nil {
			return nil, err
		}

		client.Jar = cookieJar
	}

	return client, nil
}

// Inject new haders into each request.
type addHeader struct {
	headers map[string]string
	rt      http.RoundTripper
}

// Allows to add header to each request done by this HTTP client.
func (h *addHeader) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.headers {
		req.Header.Add(k, v)
	}

	return h.rt.RoundTrip(req)
}
