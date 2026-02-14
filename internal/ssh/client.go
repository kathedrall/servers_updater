package ssh

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"servers_updater/internal/domain"
	"strings"
	"time"

	"github.com/kevinburke/ssh_config"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Client represents an SSH client connection
type Client struct {
	Client *ssh.Client
}

// Wrapper wraps an SSH client connection
type Wrapper struct {
	Client *ssh.Client
}

// ProxyWrapper wraps an SSH connection through a proxy jump
type ProxyWrapper struct {
	TargetClient *ssh.Client
	JumpClient   domain.SSHClient
}

// ExecuteCommand executes a command via SSH and returns the output
func (w *Wrapper) ExecuteCommand(cmd string) (string, error) {
	session, err := w.Client.NewSession()
	if err != nil {
		return " ", err
	}
	defer session.Close()
	out, err := session.CombinedOutput(cmd)

	return string(out), err
}

// Close closes the SSH connection
func (w *Wrapper) Close() error {
	return w.Client.Close()
}

// ExecuteCommand executes a command via SSH through proxy and returns the output
func (p *ProxyWrapper) ExecuteCommand(cmd string) (string, error) {
	session, err := p.TargetClient.NewSession()
	if err != nil {
		return " ", err
	}
	defer session.Close()
	out, err := session.CombinedOutput(cmd)

	return string(out), err
}

// Close closes both target and jump SSH connections
func (p *ProxyWrapper) Close() error {
	// Fechar conexão do target primeiro
	if p.TargetClient != nil {
		p.TargetClient.Close()
	}
	// Depois fechar conexão do jump host
	if p.JumpClient != nil {
		p.JumpClient.Close()
	}
	return nil
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

	agentClient := agent.NewClient(sock)
	signers, err := agentClient.Signers()
	if err != nil || len(signers) == 0 {
		return nil
	}

	return ssh.PublicKeys(signers...)
}

func tryPrivateKeyFile(path string) ssh.AuthMethod {
	if path == "" {
		return nil
	}

	key, err := os.ReadFile(path)
	if err != nil {
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

// Connect establishes an SSH connection to the given machine
func Connect(m domain.Machine) (domain.SSHClient, error) {
	resolveHostConfig(&m)

	// Se tem ProxyJump, usar conexão via proxy
	if m.ProxyJumper != nil {
		return connectWithProxyJump(m)
	}

	// Conexão direta normal
	return connectDirect(m)
}

// connectDirect estabelece conexão SSH direta
func connectDirect(m domain.Machine) (domain.SSHClient, error) {
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

	return &Wrapper{Client: client}, nil
}

// connectWithProxyJump estabelece conexão SSH via ProxyJump
func connectWithProxyJump(target domain.Machine) (domain.SSHClient, error) {
	// 1. Resolver configuração do jump host
	jumpMachine := domain.Machine{
		ID:   *target.ProxyJumper,
		Host: *target.ProxyJumper,
	}
	resolveHostConfig(&jumpMachine)

	// 2. Conectar no jump host
	jumpClient, err := connectDirect(jumpMachine)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to jump host %s: %v", *target.ProxyJumper, err)
	}

	// 3. Criar tunnel através do jump host
	targetAddr := formatAddress(target.Host, target.Port)
	conn, err := jumpClient.(*Wrapper).Client.Dial("tcp", targetAddr)
	if err != nil {
		jumpClient.Close()
		return nil, fmt.Errorf("failed to dial target %s through jump host: %v", targetAddr, err)
	}

	// 4. Configurar autenticação para o target
	authMethods, err := getAuthMethods(target)
	if err != nil {
		conn.Close()
		jumpClient.Close()
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            target.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	// 5. Estabelecer conexão SSH no target através do tunnel
	clientConn, chans, reqs, err := ssh.NewClientConn(conn, targetAddr, config)
	if err != nil {
		conn.Close()
		jumpClient.Close()
		return nil, fmt.Errorf("failed to establish SSH connection to target: %v", err)
	}

	// 6. Criar cliente SSH para o target
	targetClient := ssh.NewClient(clientConn, chans, reqs)

	// 7. Retornar wrapper que gerencia ambas as conexões
	return &ProxyWrapper{
		TargetClient: targetClient,
		JumpClient:   jumpClient,
	}, nil
}

// LoadMachinesFromSSHConfig loads machine configurations from SSH config file
func LoadMachinesFromSSHConfig() ([]domain.Machine, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("erro ao obter diretório home: %v", err)
	}

	configPath := filepath.Join(home, ".ssh", "config")
	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir %s: %v", configPath, err)
	}
	defer f.Close()

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar arquivo SSH config: %v", err)
	}

	var machines []domain.Machine
	hosts := cfg.Hosts

	for _, host := range hosts {
		for _, pattern := range host.Patterns {
			if pattern.String() != "*" && !strings.Contains(pattern.String(), "*") {
				machine := domain.Machine{
					ID:     pattern.String(),
					Host:   pattern.String(),
					Status: "UNKNOWN",
				}
				resolveHostConfig(&machine)
				if machine.User == "" {
					if currentUser := os.Getenv("User"); currentUser != "" {
						machine.User = currentUser
					}
				}
				if machine.Port == 0 {
					defaultPort := 22
					machine.Port = defaultPort
				}
				machines = append(machines, machine)
				break
			}
		}
	}

	return machines, nil
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

	if proxyJump, _ := cfg.Get(m.Host, "ProxyJump"); proxyJump != "" {
		m.ProxyJumper = &proxyJump
	}
	if m.KeyPath == "" {
		keyFile, _ := cfg.Get(m.Host, "IdentityFile")
		if keyFile != "" && keyFile != "~/.ssh/identity" {
			// Expandir ~ para home directory
			if strings.HasPrefix(keyFile, "~/") {
				keyFile = filepath.Join(home, keyFile[2:])
			} else if strings.HasPrefix(keyFile, "~\\") {
				keyFile = filepath.Join(home, keyFile[2:])
			}
			m.KeyPath = keyFile
		}
	}

	if m.Port == 0 {
		portStr, _ := cfg.Get(m.Host, "Port")
		if portStr != "" {
			var port int
			if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil {
				m.Port = port
			}
		}
	}
	if m.User == "" {
		user, _ := cfg.Get(m.Host, "User")
		if user != "" {
			m.User = user
		}
	}

	realHost, _ := cfg.Get(m.Host, "Hostname")
	if realHost != "" {
		m.Host = realHost
	}

	// Detectar ProxyJump

}

// NewClient creates a new SSH client with the given parameters
func NewClient(user string, host string, port string, keyPath string, password string) (*Client, error) {
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
		e := fmt.Errorf("no authentication method was provided (key or password)")
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

	return &Client{
		Client: client,
	}, nil

}
