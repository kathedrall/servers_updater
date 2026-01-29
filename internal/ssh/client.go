package ssh

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"servers_updater/internal/domain"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	Client *ssh.Client
}
type SSHWrapper struct {
	Client *ssh.Client
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

func Connect(m domain.Machine) (domain.SSHClient, error) {
	key, err := os.ReadFile(m.KeyPath)
	if err != nil {
		e := fmt.Errorf("The SSH key could not be read.: %v", err)
		return nil, e
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		e := fmt.Errorf("Invalid private key: %v", err)
		return nil, e
	}

	config := &ssh.ClientConfig{
		User: m.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", m.Host+":22", config)
	if err != nil {
		return nil, err
	}
	return &SSHWrapper{Client: client}, nil
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
