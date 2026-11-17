package serve

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

// AuthSource identifies where a request's bearer credential came from. The
// middleware tries these sources in order and uses the first one present:
//
//	AuthHeader  Authorization: Bearer <token>
//	AuthQuery   ?token=<token> (WebSocket clients that cannot set headers)
//	AuthCookie  tv_token cookie (the bundled SPA)
//
// Header wins over query and query wins over cookie, so a stale cookie cannot
// downgrade an explicit credential. AuthNone means no credential was supplied.
type AuthSource int

const (
	AuthNone AuthSource = iota
	AuthHeader
	AuthQuery
	AuthCookie
)

func (s AuthSource) String() string {
	switch s {
	case AuthHeader:
		return "header"
	case AuthQuery:
		return "query"
	case AuthCookie:
		return "cookie"
	default:
		return "none"
	}
}

// bearerCredential returns the normalized "Bearer <token>" credential and the
// source it came from. Query and cookie values are normalized to the header
// form so authMiddleware has a single comparison.
func bearerCredential(r *http.Request) (string, AuthSource) {
	if got := r.Header.Get("Authorization"); got != "" {
		return got, AuthHeader
	}
	if t := r.URL.Query().Get("token"); t != "" {
		return "Bearer " + t, AuthQuery
	}
	if c, err := r.Cookie("tv_token"); err == nil {
		return "Bearer " + c.Value, AuthCookie
	}
	return "", AuthNone
}

// OriginClass is the trust class of a request's Origin header.
//
//	OriginNone          no Origin (CLI, agents): allowed, never reflected
//	OriginLoopback      localhost, 127.0.0.0/8, or [::1]
//	OriginConfigured    an explicit --cors/config origin
//	OriginSameOrigin    same host as the request, host is an IP or localhost
//	OriginRejectedDNS   same host, but a DNS name (DNS rebinding); rejected
//	OriginRejected      a different host or an unparseable origin
//
// Loopback and configured origins are echoed in CORS headers. None, loopback,
// configured, and same-origin requests are trusted to drive the API; DNS and
// cross-origin requests are not. Trusted() models the origin-only decision, so
// a request with no Origin is trusted (non-browser clients are unaffected).
type OriginClass int

const (
	OriginNone OriginClass = iota
	OriginLoopback
	OriginConfigured
	OriginSameOrigin
	OriginRejectedDNS
	OriginRejected
)

func (c OriginClass) String() string {
	switch c {
	case OriginLoopback:
		return "loopback"
	case OriginConfigured:
		return "configured"
	case OriginSameOrigin:
		return "same-origin"
	case OriginRejectedDNS:
		return "rejected-dns"
	case OriginRejected:
		return "rejected"
	default:
		return "none"
	}
}

// Trusted reports whether the origin may drive the API (mutations, WebSocket).
func (c OriginClass) Trusted() bool {
	switch c {
	case OriginNone, OriginLoopback, OriginConfigured, OriginSameOrigin:
		return true
	default:
		return false
	}
}

// Reflect reports whether the origin may be echoed in CORS headers. Same-origin
// requests do not need the header, and untrusted origins never receive it.
func (c OriginClass) Reflect() bool {
	return c == OriginLoopback || c == OriginConfigured
}

// classifyOrigin is the single origin decision used by the CORS middleware and
// the WebSocket upgrader. Configured origins are checked before loopback so an
// explicit allow-list entry is reported as such, then same-origin IPs, then the
// DNS-rebinding rejection.
func (s *Server) classifyOrigin(r *http.Request) OriginClass {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return OriginNone
	}
	trimmed := strings.TrimRight(origin, "/")
	for _, o := range s.corsOrigins {
		if strings.EqualFold(strings.TrimRight(o, "/"), trimmed) {
			return OriginConfigured
		}
	}
	if isLoopbackOrigin(origin) {
		return OriginLoopback
	}
	u, err := url.Parse(origin)
	if err != nil || !strings.EqualFold(u.Host, r.Host) {
		return OriginRejected
	}
	host := u.Hostname()
	if strings.EqualFold(host, "localhost") || net.ParseIP(host) != nil {
		return OriginSameOrigin
	}
	return OriginRejectedDNS
}

func isLoopbackOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	host := u.Hostname()
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
