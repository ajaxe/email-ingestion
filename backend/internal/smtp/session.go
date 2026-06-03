package smtp

import (
	"context"
	"io"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/ajaxe/email-ingestion/pkg/database/public"
	"github.com/dgraph-io/ristretto"
	"github.com/emersion/go-smtp"
	"github.com/jhillyerd/enmime"
	"github.com/mileusna/spf"
)

type IngestSession struct {
	From string
	To   []string
	// ReferenceTokens maps the full recipient email to the sub-address token (if any). For example, "
	// "user+tag@example.com" -> "tag"
	ReferenceTokens map[string]string
	RemoteAddr      string
	SessionID       string
	ConnectedAt     time.Time
	cfg             *config.AppConfig
	queries         *public.Queries
	cache           *ristretto.Cache
}

func (s *IngestSession) Mail(from string, opts *smtp.MailOptions) error {
	slog.Info("Mail from", "from_email", from)

	host, _, err := net.SplitHostPort(s.RemoteAddr)
	if err != nil {
		host = s.RemoteAddr // Fallback
	}

	ip := net.ParseIP(host)
	if ip != nil {
		result := spf.CheckHost(ip, s.cfg.Smtp.Domain, from, "")
		if result == spf.Fail {
			slog.Info("SPF Validation failed", "from", from, "ip", host, "result", result)
			return &smtp.SMTPError{
				Code:         550,
				EnhancedCode: smtp.EnhancedCode{5, 7, 1},
				Message:      "SPF Validation Failed",
			}
		}
	}

	s.From = from
	return nil
}

func (s *IngestSession) Rcpt(to string, opts *smtp.RcptOptions) error {
	slog.Info("Rcpt to", "to_email", to)

	// SECURITY: Only accept emails destined for your managed domain.
	if !strings.HasSuffix(to, "@"+s.cfg.Smtp.EmailDomain) {
		slog.Info("Cannot accept email", "to_email", to)
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 1, 1},
			Message:      "User not local; please try a different gateway.",
		}
	}

	parts := strings.Split(to, "@")
	if len(parts) != 2 {
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 1, 1},
			Message:      "Invalid email address.",
		}
	}
	localPart := parts[0]

	baseLocalPart := localPart
	subAddress := ""
	if plusIdx := strings.Index(localPart, "+"); plusIdx != -1 {
		baseLocalPart = localPart[:plusIdx]
		subAddress = localPart[plusIdx+1:]
	}

	// Check cache
	var isValid bool
	if val, found := s.cache.Get(baseLocalPart); found {
		isValid = val.(bool)
	} else {
		// Cache miss, check db
		ctx := context.Background()
		_, err := s.queries.GetAssignedEmailByLocalPart(ctx, baseLocalPart)
		if err != nil {
			// Not found or db error
			isValid = false
			s.cache.SetWithTTL(baseLocalPart, false, 1, 5*time.Minute)
		} else {
			isValid = true
			s.cache.SetWithTTL(baseLocalPart, true, 1, 1*time.Hour)
		}
	}

	if !isValid {
		slog.Info("User unknown", "local_part", baseLocalPart)
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 1, 1},
			Message:      "User Unknown",
		}
	}

	s.To = append(s.To, to)
	if subAddress != "" {
		s.ReferenceTokens[to] = subAddress
	}

	return nil
}

// Data handles the 'DATA' phase where the email payload arrives.
func (s *IngestSession) Data(r io.Reader) error {
	slog.Info("Streaming incoming message payload...")

	// SECURITY: Limit the maximum size to read to prevent memory exhaustion attacks.
	// Let's say a strict 10MB ceiling limit here as a fallback.
	lr := io.LimitReader(r, s.cfg.Smtp.EmailMaxSizeBytes()) // Using Server.Port as a placeholder for max size in MB, adjust as needed.

	// Process the email stream.
	// For production, parse this using an email parser package like `github.com/emersion/go-message`
	// Here, we just print it to standard out as a proof of concept.
	/* buf := new(strings.Builder)
	if _, err := io.Copy(buf, lr); err != nil {
		return err
	} */
	envelope, err := enmime.ReadEnvelope(lr)
	if err != nil {
		slog.Info("Parser Error", "error", err)
		// Returning a 554 tells the sending server the transaction failed due to malformed data
		return &smtp.SMTPError{
			Code:         554,
			EnhancedCode: smtp.EnhancedCode{5, 6, 0},
			Message:      "Error: Failed to parse MIME topology.",
		}
	}

	slog.Info("--- SUCCESS: PARSED EMAIL ---")
	slog.Info("Subject", "subject", envelope.GetHeader("Subject"))
	slog.Info("From", "from", envelope.GetHeader("From"))
	slog.Info("Text Body", "text_body", envelope.Text)

	if len(envelope.Attachments) > 0 {
		slog.Info("Attachments detected", "count", len(envelope.Attachments))
	}
	slog.Info("-----------------------------")
	return nil
}

// Reset clears the session state (called if client sends RSET)
func (s *IngestSession) Reset() {
	s.From = ""
	s.To = nil
}

// Logout is triggered when connection is terminated.
func (s *IngestSession) Logout() error {
	return nil
}
