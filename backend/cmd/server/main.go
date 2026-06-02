package main

import (
	"io"
	"log"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/jhillyerd/enmime"
)

const emailDomainSuffix = "@gmail.com"
const maxEmailSize = 10 * 1024 * 1024 // 10MB
const maxLineLength = 2048            // Enforce a reasonable line length limit

type IngestBackend struct {
}

type IngestSession struct {
	From string
	To   []string
}

func (b *IngestBackend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	log.Printf("New connection from: %s", c.Conn().RemoteAddr().String())
	return &IngestSession{}, nil
}

func (s *IngestSession) Mail(from string, opts *smtp.MailOptions) error {
	log.Printf("Mail from: %s", from)
	s.From = from
	return nil
}

func (s *IngestSession) Rcpt(to string, opts *smtp.RcptOptions) error {
	log.Printf("Rcpt to: %s", to)

	// SECURITY: Only accept emails destined for your managed domain.
	// Prevents your server from being used as an open relay for spamming others.
	if !strings.HasSuffix(to, emailDomainSuffix) {
		log.Printf("Cannot accept email for: %s", to)
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
	log.Println("Streaming incoming message payload...")

	// SECURITY: Limit the maximum size to read to prevent memory exhaustion attacks.
	// Let's say a strict 10MB ceiling limit here as a fallback.
	lr := io.LimitReader(r, maxEmailSize)

	// Process the email stream.
	// For production, parse this using an email parser package like `github.com/emersion/go-message`
	// Here, we just print it to standard out as a proof of concept.
	/* buf := new(strings.Builder)
	if _, err := io.Copy(buf, lr); err != nil {
		return err
	} */
	envelope, err := enmime.ReadEnvelope(lr)
	if err != nil {
		log.Printf("Parser Error: %v", err)
		// Returning a 554 tells the sending server the transaction failed due to malformed data
		return &smtp.SMTPError{
			Code:         554,
			EnhancedCode: smtp.EnhancedCode{5, 6, 0},
			Message:      "Error: Failed to parse MIME topology.",
		}
	}

	log.Println("--- SUCCESS: PARSED EMAIL ---")
	log.Printf("Subject: %s", envelope.GetHeader("Subject"))
	log.Printf("From:    %s", envelope.GetHeader("From"))
	log.Printf("Text Body:\n%s", envelope.Text)

	if len(envelope.Attachments) > 0 {
		log.Printf("Attachments detected: %d files", len(envelope.Attachments))
	}
	log.Println("-----------------------------")
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

func main() {
	be := &IngestBackend{}

	s := smtp.NewServer(be)

	// --- SECURITY HARDENING SETTINGS ---
	s.Addr = "127.0.0.1:2525" // Local port for testing. In prod, use ":25"
	s.Domain = "mx.yourdomain.com"

	// Tight timeouts protect against slow TCP resource starvation attacks
	s.ReadTimeout = 10 * time.Second
	s.WriteTimeout = 10 * time.Second

	// Hard payload size cap enforced at the protocol level (e.g., 5MB limit)
	s.MaxMessageBytes = maxEmailSize

	// Enforce a sensible line length limit to block buffer overflow exploits
	s.MaxLineLength = maxLineLength

	// Disable authentication since this is an ingestion-only public server.
	// Public MX servers must accept unauthenticated mail from foreign MTAs.
	s.AllowInsecureAuth = false

	// For production, you MUST provide TLS configurations:
	// cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	// if err == nil {
	//     s.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	//     s.AllowInsecureAuth = false
	// }

	log.Printf("Starting secure receive-only SMTP server on %s", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Server structural failure: %v", err)
	}
}
