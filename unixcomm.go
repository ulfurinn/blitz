package blitz

import (
	"net"
	"net/http"
	"strings"
	"time"
)

type unixDialer struct {
	net.Dialer
}

func (d *unixDialer) Dial(network, address string) (net.Conn, error) {
	parts := strings.Split(address, ":")
	return d.Dialer.Dial("unix", parts[0])
}

var unixTransport http.RoundTripper = &http.Transport{
	Dial: (&unixDialer{net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	},
	}).Dial,
}
