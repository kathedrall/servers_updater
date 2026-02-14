package pkgmanager

import (
	"fmt"
	"servers_updater/internal/domain"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const releaseName = "cat /etc/os-release || uname -a"

var (
	CommandCheckUpdate = [7]string{
		"sudo dnf check-update",
		"sudo apk list -u",
		"sudo apt-get upgrade -s",
		"sudo zypper list-updates",
		"sudo pacman -Sy && pacman -Qu",
		"sudo emerge -pvu @world",
		"sudo pkg upgrade -n",
	}
	CommandUpdate = [7]string{
		"sudo dnf update -y",
		"sudo apk upgrade --no-cache",
		"sudo DEBIAN_FRONTEND=noninteractive apt-get upgrade -y",
		"sudo zypper update --non-interactive",
		"sudo pacman -Syu --noconfirm",
		"sudo emerge -u --auto-umask-write @world",
		"sudo pkg upgrade -y",
	}
	CommandCountPackages = [7]string{
		`sudo dnf check-update --quiet | grep -v '^$' | wc -l`,
		`sudo apk list -u | wc -l`,
		`sudo apt-get -s -o Debug::NoLocking=true upgrade | grep -c ^Inst`,
		`sudo zypper -q lu --best-effort | grep -c 'v |'`,
		`sudo checkupdates | wc -l`,
		`sudo emerge -pvuND @world | grep -c "ebuild"`,
		`pkg version -l "<" | wc -l`,
	}
	CommandRestartCheck = [7]string{
		` needs-restarting -r > /dev/null 2>&1; if [ $? -eq 1]; then echo "SIM"; else echo "NAO"; fi`,
		` if [ "$(uname -r)" != $(cat /proc/sys/kernel/osrelease 2>/dev/null) ]; then echo "SIM"; else echo "NAO"; fi`,
		`[ -f /var/run/reboot-required ] && echo '"SIM"' || '"NAO"`,
		`sudo zypper ps -s >/dev/null 2>&1 && echo "NAO" || echo "SIM"`,
		`running=$(uname -r); installed=$(pacman -Q linux | awk '{print $2}'); if [[ "$running" != *"$installed"* ]]; then echo "SIM"; else echo "NAO"; fi`,
		`if [ "$(uname -r)" != "$(eselect kernel show | grep -o 'linux-[0-9].*')" ]; then echo "SIM"; else echo "NAO"; fi`,
		`if [ "$(freebsd-version -k)" != "$(uname -r)" ]; then echo "SIM"; else echo "NAO"; fi`,
	}
)

type PackageManager interface {
	GetCheckCommand() string
	GetInstallCommand() string
	ParseOutput(output string) ([]domain.Package, error)
}

// DetectOs Conecta na maquina e descobre o sistema operacional
func DetectOs(client domain.SSHClient, m *domain.Machine) error {
	output, err := client.ExecuteCommand(releaseName)
	if err != nil {
		e := fmt.Errorf("failed to detect OS: %v", err)
		return e
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

	if m.OSName == "" {
		if strings.Contains(strings.ToLower(output), "freebsd") {
			m.PrettyName = "freebsd"
			m.PrettyName = "FreeBSD"
		} else {
			m.OSName = "unknown"
		}
	}

	if m.PrettyName == "" {
		m.PrettyName = cases.Title(language.Und).String(m.OSName)
	}

	return nil
}

func GetManager(osName string) (PackageManager, error) {
	switch osName {
	case "ubuntu", "debian", "kali", "pop", "linuxmint":
		return &AptManager{}, nil
	case "fedora", "centos", "rhel", "rocky", "almalinux":
		return &DnfManager{}, nil
	case "alpine":
		return &ApkManager{}, nil
	case "opensuse", "opensuse-leap", "opensuse-tumbleweed", "sles":
		return &ZypperManager{}, nil
	case "arch", "manjaro":
		return &PacmanManager{}, nil
	case "gentoo":
		return &EmergeManager{}, nil
	case "freebsd":
		return &FreebsdManager{}, nil
	default:
		e := fmt.Errorf("OS not supported: %s", osName)
		return nil, e
	}
}
