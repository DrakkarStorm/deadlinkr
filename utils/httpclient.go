package utils

import (
	"net"
	"net/http"
	"time"

	"github.com/DrakkarStorm/deadlinkr/model"
)

var (
	// ClientHTTP is the shared client for all requests
	ClientHTTP = &http.Client{
		Transport: &http.Transport{
			// Maximum number of idle connections to keep open
			MaxIdleConns: 400,
			// Maximum number of idle connections to keep open per host
			MaxIdleConnsPerHost: 200,
			// Delay before closing an idle connection
			IdleConnTimeout: 90 * time.Second,
			// Timeout of dial (TCP connection establishment)
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			// Timeout TLS handshake if HTTPS
			TLSHandshakeTimeout: 10 * time.Second,
		},
		// Timeout global for the entire request (connect + headers + body)
		Timeout: time.Duration(model.Timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Return nil to follow redirects
			return nil
		},
	}
)
