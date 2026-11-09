// Package localserve exposes a local media file over HTTP so an Android TV
// player can stream it.
package localserve

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"
)

// Server is a running file server for a single media file.
type Server struct {
	File     string
	URL      string
	listener net.Listener
	http     *http.Server
}

// NewHandler serves exactly one file at /<base>. Every other path is 404.
func NewHandler(file string) http.Handler {
	base := "/" + filepath.Base(file)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != base {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, file)
	})
}

// Start begins serving file on all interfaces and returns once the listener is
// bound. Pass port 0 to choose a free port. The advertise IP is derived from
// tvIP (the TV address the client can reach), falling back to 127.0.0.1.
func Start(file, tvIP string, port int) (*Server, error) {
	ln, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(port))
	if err != nil {
		return nil, err
	}

	actualPort := ln.Addr().(*net.TCPAddr).Port
	ip := AdvertiseIP(tvIP)
	u := "http://" + net.JoinHostPort(ip, strconv.Itoa(actualPort)) + "/" + url.PathEscape(filepath.Base(file))

	s := &Server{
		File:     file,
		URL:      u,
		listener: ln,
		http:     &http.Server{Handler: NewHandler(file)},
	}
	go func() {
		_ = s.http.Serve(ln)
	}()
	return s, nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// AdvertiseIP returns the local address that can route to tvIP. tvIP may be a
// bare host or host:port. The address is discovered without sending packets.
func AdvertiseIP(tvIP string) string {
	host := tvIP
	if h, _, err := net.SplitHostPort(tvIP); err == nil {
		host = h
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		return "127.0.0.1"
	}
	conn, err := net.DialTimeout("udp", net.JoinHostPort(host, "9"), time.Second)
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok && addr.IP != nil && !addr.IP.IsUnspecified() {
		return addr.IP.String()
	}
	return "127.0.0.1"
}
