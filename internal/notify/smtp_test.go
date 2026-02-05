package notify

import (
	"os"
	"servers_updater/internal/domain"
	"testing"
	"time"
)

func TestSendHtmlEmail_Validation(t *testing.T) {
 config := domain.SMTPConfig{
  Host: "smtp.gmail.com",
  Port: "587",
 }
 data := domain.EmailData{
  Host: "TestBox",
 }

 err := SendHtmlEmail(config, []string{}, data)
 if err == nil {
  t.Error("Expected error for empty recepients list, got nil")
 } else if err.Error() != "no recipients provided" {
  t.Errorf("Unexpected error message: %v", err)
 }
}

func TestSendHtmlEmail_Integration_REAL(t *testing.T) {
  host := os.Getenv("TEST_SMTP_HOST")
  user := os.Getenv("TEST_SMTP_USER")
  pass := os.Getenv("TEST_SMTP_PASS")
  to   := os.Getenv("TEST_SMTP_TO")

  if host == "" || user == "" || pass == "" || to == "" {
   t.Skip("Skipping integration test: enviroment variables not set (TEST_SMTP_HOST, etc)")
 }

 config := domain.SMTPConfig {
  Host: host,
  Port: "587",
  User: user,
  Password: pass,
 }

 now := time.Now()
 data := domain.EmailData{
  Host:   "INTEGRATION-TEST-RUNNER",
  Status: "TESTING",
  Date:   &now,
  ErrorMessage: "This is a test email triggered by 'go test'",
  Packages: []domain.Package{
   {
     Name: "git-test",
     NewVersion: "1.0", 
     CurrentVersion: "0.1",
   },
  },
 }
 
 err := SendHtmlEmail(config, []string{to}, data)
 if err != nil {
  t.Errorf("Falied to send real email: %v", err)
 }
}

