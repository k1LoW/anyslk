package smtp_server

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/k1LoW/anyslk/util"
	slack "github.com/monochromegane/slack-incoming-webhooks"
	"go.uber.org/zap"
)

// The Backend implements SMTP server methods.
type Backend struct {
	webhookURL string
	logger     *zap.Logger
}

// Login handles a login command with username and password.
func (be *Backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	return &Session{
		username:   username,
		webhookURL: be.webhookURL,
		logger:     be.logger,
	}, nil
}

// AnonymousLogin requires clients to authenticate using SMTP AUTH before sending emails
func (be *Backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	return &Session{
		username:   "anyslk",
		webhookURL: be.webhookURL,
		logger:     be.logger,
	}, nil
}

// A Session is returned after successful login.
type Session struct {
	webhookURL string
	logger     *zap.Logger
	username   string
	from       string
	to         string
}

func (s *Session) Mail(from string) error {
	s.from = from
	return nil
}

func (s *Session) Rcpt(to string) error {
	s.to = to
	return nil
}

func (s *Session) Data(r io.Reader) error {
	m, err := mail.ReadMessage(r)
	if err != nil {
		return err
	}
	header := m.Header
	subject := header.Get("Subject")
	from := header.Get("From")
	to := header.Get("To")
	date := header.Get("Date")
	body, err := ioutil.ReadAll(m.Body)
	if err != nil {
		return err
	}
	if subject == "" {
		subject = fmt.Sprintf("Mail from %s", s.from)
	}
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	slackChannel := makeSlackChannel(s.to)
	payload := slack.Payload{
		Channel:   slackChannel,
		IconEmoji: ":slack:",
		Username:  s.username,
	}
	attachment := slack.Attachment{
		Title:     fmt.Sprintf(":email: %s", subject),
		Fallback:  subject,
		Timestamp: time.Now().Unix(),
		Footer:    fmt.Sprintf("anyslk on %s", hostname),
	}
	attachment.AddField(&slack.Field{
		Title: "From",
		Value: from,
		Short: true,
	})
	attachment.AddField(&slack.Field{
		Title: "To",
		Value: to,
		Short: true,
	})
	attachment.AddField(&slack.Field{
		Title: "Date",
		Value: date,
		Short: true,
	})
	attachment.AddField(&slack.Field{
		Title: "Body",
		Value: string(body),
		Short: false,
	})

	payload.AddAttachment(&attachment)
	slack.Client{
		WebhookURL: s.webhookURL,
	}.Post(&payload)
	s.logger.Info("Mail -> slack message", zap.String("from", s.from), zap.String("to", s.to), zap.Any("header", m.Header), zap.String("body", string(body)))
	return nil
}

func (s *Session) Reset() {}

func (s *Session) Logout() error {
	return nil
}

// makeSlackChannel ...
func makeSlackChannel(to string) string {
	splitted := strings.Split(to, "@")
	return fmt.Sprintf("#%s", splitted[0])
}

// Run ...
func Run(ctx context.Context, logger *zap.Logger, port int) error {
	webhookURL, err := util.GetEnvSlackIncommingWebhook()
	if err != nil {
		return err
	}
	be := &Backend{
		webhookURL: webhookURL,
		logger:     logger,
	}
	s := smtp.NewServer(be)
	defer s.Close()
	s.Addr = fmt.Sprintf("localhost:%d", port)
	s.Domain = "anyslk.local"
	s.ReadTimeout = 1000 * time.Second
	s.WriteTimeout = 1000 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	logger.Info(fmt.Sprintf("Start listening %s", s.Addr))

	go func() {
		if err := s.ListenAndServe(); err != nil {
			logger.Fatal("error", zap.Error(err))
		}
	}()

	select {
	case <-ctx.Done():
		break
	}
	return nil
}

// RunWithServerStarter ...
func RunWithServerStarter(ctx context.Context, logger *zap.Logger, l net.Listener) error {
	webhookURL, err := util.GetEnvSlackIncommingWebhook()
	if err != nil {
		return err
	}
	be := &Backend{
		webhookURL: webhookURL,
		logger:     logger,
	}
	s := smtp.NewServer(be)
	defer s.Close()
	s.Addr = l.Addr().String()
	s.Domain = "anyslk.local"
	s.ReadTimeout = 1000 * time.Second
	s.WriteTimeout = 1000 * time.Second
	s.MaxMessageBytes = 1024 * 1024
	s.MaxRecipients = 50
	s.AllowInsecureAuth = true

	logger.Info(fmt.Sprintf("Start listening %s", s.Addr))

	go func() {
		if err := s.Serve(l); err != nil {
			logger.Fatal("error", zap.Error(err))
		}
	}()

	select {
	case <-ctx.Done():
		break
	}
	return nil
}
