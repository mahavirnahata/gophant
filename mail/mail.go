// Package mail provides a fluent email sending API with swappable drivers.
//
// Usage:
//
//	mailer := mail.New(mail.NewSMTPDriver(mail.SMTPConfig{...}))
//	mailer.To("alice@example.com").
//	    Subject("Welcome!").
//	    HTML("<h1>Hi Alice</h1>").
//	    Text("Hi Alice").
//	    Send()
package mail

import (
	"fmt"
	"log"
	"mime"
	"net/smtp"
	"strings"
)

// Message is a fully-assembled email message.
type Message struct {
	From    string
	ReplyTo string
	To      []string
	CC      []string
	BCC     []string
	Subject string
	Text    string
	HTML    string
}

// Driver sends messages.
type Driver interface {
	Send(msg *Message) error
}

// Mailer is the entry point for building and dispatching emails.
type Mailer struct {
	driver Driver
	From   string
}

// New returns a Mailer backed by the given driver.
func New(driver Driver) *Mailer {
	return &Mailer{driver: driver}
}

// To starts building a message addressed to the given recipients.
func (m *Mailer) To(addrs ...string) *PendingMail {
	return &PendingMail{mailer: m, msg: Message{From: m.From, To: addrs}}
}

// Send is a shortcut for m.To(to).Subject(subj).Text(body).Send().
func (m *Mailer) Send(to, subject, body string) error {
	return m.To(to).Subject(subject).Text(body).Send()
}

// PendingMail builds a single email before sending.
type PendingMail struct {
	mailer *Mailer
	msg    Message
}

func (p *PendingMail) From(addr string) *PendingMail    { p.msg.From = addr; return p }
func (p *PendingMail) ReplyTo(addr string) *PendingMail { p.msg.ReplyTo = addr; return p }
func (p *PendingMail) CC(addrs ...string) *PendingMail  { p.msg.CC = addrs; return p }
func (p *PendingMail) BCC(addrs ...string) *PendingMail { p.msg.BCC = addrs; return p }
func (p *PendingMail) Subject(s string) *PendingMail    { p.msg.Subject = s; return p }
func (p *PendingMail) Text(body string) *PendingMail    { p.msg.Text = body; return p }
func (p *PendingMail) HTML(html string) *PendingMail    { p.msg.HTML = html; return p }

// Send dispatches the message via the underlying driver.
func (p *PendingMail) Send() error {
	return p.mailer.driver.Send(&p.msg)
}

// ── SMTP Driver ───────────────────────────────────────────────────────────────

// SMTPConfig holds SMTP connection settings.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SMTPDriver struct{ cfg SMTPConfig }

// NewSMTPDriver creates a driver that sends via net/smtp.
func NewSMTPDriver(cfg SMTPConfig) *SMTPDriver { return &SMTPDriver{cfg: cfg} }

func (d *SMTPDriver) Send(msg *Message) error {
	from := msg.From
	if from == "" {
		from = d.cfg.From
	}
	addr := fmt.Sprintf("%s:%d", d.cfg.Host, d.cfg.Port)

	var auth smtp.Auth
	if d.cfg.Username != "" {
		auth = smtp.PlainAuth("", d.cfg.Username, d.cfg.Password, d.cfg.Host)
	}

	recipients := append(append([]string{}, msg.To...), msg.CC...)
	recipients = append(recipients, msg.BCC...)
	return smtp.SendMail(addr, auth, from, recipients, []byte(buildRFC2822(from, msg)))
}

func buildRFC2822(from string, msg *Message) string {
	var sb strings.Builder
	header := func(k, v string) { fmt.Fprintf(&sb, "%s: %s\r\n", k, v) }

	header("From", from)
	header("To", strings.Join(msg.To, ", "))
	if len(msg.CC) > 0 {
		header("Cc", strings.Join(msg.CC, ", "))
	}
	if msg.ReplyTo != "" {
		header("Reply-To", msg.ReplyTo)
	}
	header("Subject", mime.QEncoding.Encode("utf-8", msg.Subject))
	header("MIME-Version", "1.0")

	switch {
	case msg.HTML != "" && msg.Text != "":
		boundary := "gophant_alt_boundary"
		header("Content-Type", fmt.Sprintf(`multipart/alternative; boundary="%s"`, boundary))
		sb.WriteString("\r\n")
		fmt.Fprintf(&sb, "--%s\r\n", boundary)
		sb.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
		sb.WriteString(msg.Text + "\r\n")
		fmt.Fprintf(&sb, "--%s\r\n", boundary)
		sb.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
		sb.WriteString(msg.HTML + "\r\n")
		fmt.Fprintf(&sb, "--%s--\r\n", boundary)
	case msg.HTML != "":
		header("Content-Type", "text/html; charset=utf-8")
		sb.WriteString("\r\n")
		sb.WriteString(msg.HTML)
	default:
		header("Content-Type", "text/plain; charset=utf-8")
		sb.WriteString("\r\n")
		sb.WriteString(msg.Text)
	}
	return sb.String()
}

// ── Log Driver ────────────────────────────────────────────────────────────────

// LogDriver prints outgoing emails to stderr (for local development).
type LogDriver struct{}

func NewLogDriver() *LogDriver { return &LogDriver{} }

func (d *LogDriver) Send(msg *Message) error {
	log.Printf("[mail] To=%v Subject=%q\n%s", msg.To, msg.Subject, msg.Text+msg.HTML)
	return nil
}

// ── Null Driver ───────────────────────────────────────────────────────────────

// NullDriver silently discards all messages and records them for inspection.
// Use in tests: driver.AssertSent(t, "alice@example.com").
type NullDriver struct {
	Sent []*Message
}

func NewNullDriver() *NullDriver { return &NullDriver{} }

func (d *NullDriver) Send(msg *Message) error {
	d.Sent = append(d.Sent, msg)
	return nil
}

// AssertSent fails the test if no message was sent to addr.
func (d *NullDriver) AssertSent(t interface{ Fatalf(string, ...any) }, addr string) {
	for _, m := range d.Sent {
		for _, to := range m.To {
			if to == addr {
				return
			}
		}
	}
	t.Fatalf("mail: expected message sent to %q, got %d total messages", addr, len(d.Sent))
}

// AssertNotSent fails the test if any message was sent.
func (d *NullDriver) AssertNotSent(t interface{ Fatalf(string, ...any) }) {
	if len(d.Sent) > 0 {
		t.Fatalf("mail: expected no messages sent, got %d", len(d.Sent))
	}
}

// Reset clears captured messages.
func (d *NullDriver) Reset() { d.Sent = nil }
