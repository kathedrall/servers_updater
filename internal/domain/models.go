package domain

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
 Server string
 Port string
 User string
 Password string
 To string
}
