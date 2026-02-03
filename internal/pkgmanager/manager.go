package pkgmanager

import (
	"fmt"
	"servers_updater/internal/domain"
)

var (
	CommandCheckUpdate = []string{
		"sudo dnf check-update",
		"sudo apk list -u",
		"sudo apt-get upgrade -s",
		"sudo zypper list-updates",
		"sudo pacman -Sy && pacman -Qu",
		"sudo emerge -pvu @world",
		"sudo pkg upgrade -n",
	}
	CommandUpdate = []string{
		"sudo dnf update -y",
		"sudo apk upgrade --no-cache",
		"sudo DEBIAN_FRONTEND=noninteractive apt-get upgrade -y",
		"sudo zypper update --non-interactive",
		"sudo pacman -Syu --noconfirm",
		"sudo emerge -u --auto-umask-write @world",
		"sudo pkg upgrade -y",
	}
)

type PackageManager interface {
	GetCheckCommand() string
	GetInstallCommand() string
	ParseOutput(output string) ([]domain.Package, error)
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
