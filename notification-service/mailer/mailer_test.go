package mailer_test

import (
	"testing"

	"github.com/mechmanager/fiapx/notification-service/mailer"
)

func TestNewMailer_EmptyHost_ReturnsLogMailer(t *testing.T) {
	m := mailer.NewMailer("", "587", "", "", "from@test.com")
	if m == nil {
		t.Fatal("expected non-nil Mailer")
	}
	if err := m.Send("to@test.com", "Assunto", "corpo do email"); err != nil {
		t.Errorf("LogMailer.Send returned unexpected error: %v", err)
	}
}

func TestNewMailer_WithHost_ReturnsSMTPMailer(t *testing.T) {
	m := mailer.NewMailer("127.0.0.1", "1", "user", "pass", "from@test.com")
	if m == nil {
		t.Fatal("expected non-nil Mailer")
	}
	err := m.Send("to@test.com", "Assunto", "corpo do email")
	if err == nil {
		t.Error("expected connection error to invalid SMTP host")
	}
}
