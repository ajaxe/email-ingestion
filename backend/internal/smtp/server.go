package smtp

import (
	"log/slog"
	"time"

	"github.com/ajaxe/email-ingestion/pkg/config"
	"github.com/emersion/go-smtp"
	"github.com/google/uuid"
)

type SmtpServerBackend struct {
	cfg *config.AppConfig
}

func NewSmtpServer(cfg *config.AppConfig) *smtp.Server {
	be := &SmtpServerBackend{cfg: cfg}

	s := smtp.NewServer(be)

	// --- SECURITY HARDENING SETTINGS ---
	s.Addr = cfg.Smtp.ListenAddress // Local port for testing. In prod, use ":25"
	s.Domain = cfg.Smtp.Domain

	// Tight timeouts protect against slow TCP resource starvation attacks
	s.ReadTimeout = time.Duration(cfg.Smtp.ReadTimeoutSec) * time.Second
	s.WriteTimeout = time.Duration(cfg.Smtp.WriteTimeoutSec) * time.Second

	// Hard payload size cap enforced at the protocol level (e.g., 5MB limit)
	s.MaxMessageBytes = cfg.Smtp.EmailMaxSizeBytes()

	// Enforce a sensible line length limit to block buffer overflow exploits
	s.MaxLineLength = cfg.Smtp.MaxLineLength

	// Disable authentication since this is an ingestion-only public server.
	// Public MX servers must accept unauthenticated mail from foreign MTAs.
	s.AllowInsecureAuth = false

	// For production, you MUST provide TLS configurations:
	// cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	// if err == nil {
	//     s.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	//     s.AllowInsecureAuth = false
	// }
	return s
}

func (b *SmtpServerBackend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	slog.Info("New connection from", "remote_addr", c.Conn().RemoteAddr().String())
	return &IngestSession{
		cfg:         b.cfg,
		RemoteAddr:  c.Conn().RemoteAddr().String(),
		SessionID:   uuid.New().String(),
		ConnectedAt: time.Now(),
	}, nil
}
