package webhook

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ajaxe/email-ingestion/pkg/config"
)

var (
	ErrPrivateIPBlocked = errors.New("webhook delivery blocked: target resolves to a private, loopback, or link-local IP address")
)

// NewSSRFProtectedClient creates an HTTP client with a custom dialer that blocks RFC 1918 IPs
// unless the domain is present in the allowedDomains list.
// It also skips the check if the application is marked as isTrusted.
func NewSSRFProtectedClient(cfg *config.WebhookConfig, isTrusted bool) *http.Client {
	// Pre-process allowed domains to lowercase for case-insensitive matching
	allowedMap := make(map[string]bool)
	if cfg != nil {
		for _, d := range cfg.AllowedDomains {
			allowedMap[strings.ToLower(strings.TrimSpace(d))] = true
		}
	}

	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			// If the tenant is trusted, skip the SSRF IP checks entirely
			if isTrusted {
				return dialer.DialContext(ctx, network, addr)
			}

			// If the exact hostname is explicitly allowed, skip the IP checks
			if allowedMap[strings.ToLower(host)] {
				return dialer.DialContext(ctx, network, addr)
			}

			// Resolve IPs for the host
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}

			if len(ips) == 0 {
				return nil, fmt.Errorf("no IPs found for host %s", host)
			}

			// Ensure all resolved IPs are public
			for _, ipAddr := range ips {
				ip := ipAddr.IP
				if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
					return nil, fmt.Errorf("%w: %s resolves to %s", ErrPrivateIPBlocked, host, ip.String())
				}
			}

			// All IPs are public, proceed with normal dialing
			return dialer.DialContext(ctx, network, addr)
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}
