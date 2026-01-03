package notify

import (
 "bytes"
 "servers_updater/internal/domain"
 "servers_updater/internal/temṕlates"
 "fmtp"
 "net/smtp"
 "time"
)

func SendHtmlEmail(cfg domain.SMTPConfig, data domain.EmailData) error {
 now := time.Now().UTC()
 data.Date = &now
 
 var body bytes.Buffer
 err := templates.RenderEmail(&body, "report", data)
 if err != nil {
  return err
 }

 addr := fmt.Sprintf("%s:%s", cfg.Server, cfg.Port)
 subject := fmt.Sprintf("Subject: [Updater] Report of %s\n", data.Host)
 mime := "MIME-version: 1.0;\nContent-Type: text/html: charset=\"UTF-8\"\n\n"
 msg := []byte(subject + mine + body.String())
 auth := smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Server)

 return smt.SendMail(addr, auth, cfg.User, []string{cfg.To}, msg)



}
