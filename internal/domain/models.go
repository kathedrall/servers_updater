package domain

import "time"

type Machine struct {
 Id string `json:"id"`
 Host string `json:"host"`
 User string `json:"user"`
 Port int `json:"port"`
 Password string `json:"password"`
 KeyPath string `json:"key_path"`
 IsVulnerable bool `json:"is_vulnerable"`
}

type Package struct {
 Name string `json:"name"`
 CurrentVersion string `json:"current_version"`
 NewVersion string `json:"new_version"`
 CVEs []string `json:"cves"`
}

type SMTPConfig struct {
 Host     string
 Port     string
 User     string
 Password string
}

type EmailData struct {
 Host string
 Status string
 Packages []Package
 ErrorMessage string
 Date *time.Time
}

type SSHClient interface {
 ExecuteCommand(cmd string) (string, error)
 Close() error
}
