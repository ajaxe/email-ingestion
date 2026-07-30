package smtp

import (
	"context"
	"io"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/ajaxe/email-ingestion/internal/service"
	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/emersion/go-smtp"
	"github.com/jhillyerd/enmime"
	"github.com/mileusna/spf"
)

type IngestSession struct {
	From         string
	To           []string
	RemoteAddr   string
	SessionID    string
	ConnectedAt  time.Time
	cfg          *config.AppConfig
	ctx          context.Context
	cancel       context.CancelFunc
	emailService *service.EmailIngestionService
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
			slog.Info("SPF Validation failed", "ip", ip, "domain", s.cfg.Smtp.Domain, "from", from, "result", result)
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

	isValid, err := s.emailService.CheckAssignedEmail(s.ctx, to)

	if !isValid {
		slog.Info("invalid inbound email", slog.String("to", to), slog.Any("error", err))
		return &smtp.SMTPError{
			Code:         550,
			EnhancedCode: smtp.EnhancedCode{5, 1, 1},
			Message:      "User Unknown",
		}
	}

	s.To = append(s.To, to)

	return nil
}

// Data handles the 'DATA' phase where the email payload arrives.
func (s *IngestSession) Data(r io.Reader) error {
	slog.Info("Streaming incoming message payload...")

	err := s.emailService.Process(s.ctx, r)

	if err != nil {
		slog.ErrorContext(s.ctx, "failed to process email", "error", err)
		return &smtp.SMTPError{
			Code:         451,
			EnhancedCode: smtp.EnhancedCode{4, 3, 0},
			Message:      "Temporary server error. Please try again later.",
		}
	}

	envelope, err := enmime.ReadEnvelope(r)
	if err != nil {
		slog.Info("email parser error", "error", err)
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
	s.cancel()
	return nil
}
