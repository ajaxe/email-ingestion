package smtp

import (
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/emersion/go-smtp"
	"github.com/jhillyerd/enmime"
)

type IngestSession struct {
	From        string
	To          []string
	RemoteAddr  string
	SessionID   string
	ConnectedAt time.Time
	cfg         *config.AppConfig
}

func (s *IngestSession) Mail(from string, opts *smtp.MailOptions) error {
	slog.Info("Mail from", "from_email", from)
	s.From = from
	return nil
}

func (s *IngestSession) Rcpt(to string, opts *smtp.RcptOptions) error {
	slog.Info("Rcpt to", "to_email", to)

	// SECURITY: Only accept emails destined for your managed domain.
	// Prevents your server from being used as an open relay for spamming others.
	if !strings.HasSuffix(to, s.cfg.Smtp.EmailDomain) {
		slog.Info("Cannot accept email", "to_email", to)
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 1, 1},
			Message:      "User not local; please try a different gateway.",
		}
	}

	s.To = append(s.To, to)
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
