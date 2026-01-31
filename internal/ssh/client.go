package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"servers_updater/internal/domain"
	"strings"
	"time"

	"github.com/kevinburke/ssh_config"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type SSHClient struct {
	Client *ssh.Client
}
type SSHWrapper struct {
	Client *ssh.Client
}

func (w *SSHWrapper) ExecuteCommand(cmd string) (string, error) {
	session, err := w.Client.NewSession()
	if err != nil {
		return " ", err
	}
	defer session.Close()
	out, err := session.CombinedOutput(cmd)

	return string(out), err
}

func (w *SSHWrapper) Close() error {
	return w.Client.Close()
}

func resolveHostConfig(m *domain.Machine) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	configPath := filepath.Join(home, ".ssh", "config")
	f, err := os.Open(configPath)
	if err != nil {
		return
	}
	defer f.Close()

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return
	}

	realHost, _ := cfg.Get(m.Host, "Hostname")
	if realHost != "" {
		m.Host = realHost
	}

	if m.User == "" {
		user, _ := cfg.Get(m.Host, "User")
		if user != "" {
			m.User = user
		}
	}

	if m.KeyPath == "" {
		keyFile, _ := cfg.Get(m.Host, "IdentityFile")
		if keyFile != "" && keyFile != "˜./.ssh/identity" {
			if strings.HasPrefix(keyFile, "˜/") {
				keyFile = filepath.Join(home, keyFile[2:])
			}
			m.KeyPath = keyFile
		}
	}

	if m.Port == 0 {
		portStr, _ := cfg.Get(m.Host, "Port")
		if portStr != "" {
			fmt.Scanf(portStr, "%d", &m.Port)
		}
	}
}

func formatAddress(host string, port int) string {
	p := "22"
	if port > 0 {
		p = fmt.Sprintf("%d", port)
	}
	return net.JoinHostPort(host, p)
}

func getAuthMethods(m domain.Machine) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod
	if method := trySSHAgent(); method != nil {
		methods = append(methods, method)
	}

	if method := tryPrivateKeyFile(m.KeyPath); method != nil {
		methods = append(methods, method)
	}

	if method := tryPassword(m.Password); method != nil {
		methods = append(methods, method)
	}

	if len(methods) == 0 {
		e := fmt.Errorf("No valid credential found for %s (no keys or password)", m.Host)
		return nil, e
	}
	return methods, nil
}

func trySSHAgent() ssh.AuthMethod {
	sockPath := os.Getenv("SSH_AUTH_SOCK")
	if sockPath == "" {
		return nil
	}

	sock, err := net.Dial("unix", sockPath)
	if err != nil {
		return nil
	}

	return ssh.PublicKeysCallback(agent.NewClient(sock).Signers)
}

func tryPrivateKeyFile(path string) ssh.AuthMethod {
	if path == "" {
		return nil
	}

	key, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Warning: The key could not be read in %s", path)
		return nil
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil
	}

	return ssh.PublicKeys(signer)
}

func tryPassword(password string) ssh.AuthMethod {
	if password == "" {
		return nil
	}

	return ssh.Password(password)
}

func Connect(m domain.Machine) (domain.SSHClient, error) {
	resolveHostConfig(&m)

	addr := formatAddress(m.Host, m.Port)

	authMethods, err := getAuthMethods(m)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            m.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		e := fmt.Errorf("Falied to connect to %s: %v", addr, err)
		return nil, e
	}

	return &SSHWrapper{Client: client}, nil
}

func NewSSHClient(user string, host string, port string, keyPath string, password string) (*SSHClient, error) {
	var authMethods []ssh.AuthMethod
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			signer, err := ssh.ParsePrivateKey(key)
			if err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	if len(authMethods) == 0 {
		e := fmt.Errorf("No authentication method was provided (key or password).")
		return nil, e
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	addr := net.JoinHostPort(host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		e := fmt.Errorf("failed to connect on %s: %v", host, err)
		return nil, e
	}

	return &SSHClient{
		Client: client,
	}, nil

}

func ParseAptOutput(output string) []domain.Package {
	var pkgs []domain.Package
	re := regexp.MustCompile(`Inst\s+([^\s]+)\s+\[([^\]]+)\]\s+\(([^\s]+)`)
	matches := re.FindAllStringSubmatch(output, -1)

	for _, m := range matches {
		pkgs = append(pkgs, domain.Package{
			Name:           m[1],
			CurrentVersion: m[2],
			NewVersion:     m[3],
		})
	}
	return pkgs
}

func IdentifyOS(client domain.SSHClient, m *domain.Machine) error {
	output, err := client.ExecuteCommand("cat /etc/os-release")
	if err != nil {
		return err
	}

	m.PrettyName = ""
	m.OSName = ""
	m.IsSupported = false

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.ReplaceAll(line, "\"", "")

		if strings.HasPrefix(line, "PRETTY_NAME=") {
			m.PrettyName = strings.TrimPrefix(line, "PRETTY_NAME=")
		}
		if strings.HasPrefix(line, "ID=") {
			m.OSName = strings.TrimPrefix(line, "ID=")
		}
	}

	if m.OSName == "debian" || m.OSName == "ubuntu" {
		m.IsSupported = true
	}

	if m.PrettyName == "" {
		if m.OSName != "" {
			m.PrettyName = cases.Title(language.Und).String(m.OSName)
		} else {
			m.PrettyName = "Generic GNU Linux"
		}
	}

	return nil
}
