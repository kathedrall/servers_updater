package notify

import (
	"bytes"
	"fmt"
	"net/smtp"
	"servers_updater/internal/domain"
	"servers_updater/internal/templates"
	"strings"
	"time"
)

func SendHtmlEmail(cfg domain.SMTPConfig, recipients []string,   data domain.EmailData) error {
 if len(recipients) == 0 {
  return fmt.Errorf("no recipients provided")
 } 

 now := time.Now().UTC()
 data.Date = &now
 
 var body bytes.Buffer
 err := templates.RenderEmail(&body, "report", data)
 if err != nil {
  return err
 }

 toHeader := strings.Join(recipients, " , ")
 addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

 headers := fmt.Sprintf("From: %s\r\n", cfg.User)
 headers += fmt.Sprintf("To: %s\r\n", toHeader)
 headers += fmt.Sprintf("Subject: [Updater] Report - %s\r\n", data.Host)
 headers += "MIME-version: 1.0;\r\n"
 headers += "Content-Type: text/html: charset=\"UTF-8\";\r\n"
 headers += "\r\n"

 msg := []byte(headers + body.String())
 auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
 
 return smtp.SendMail(addr, auth, cfg.User, recipients, msg)
}
