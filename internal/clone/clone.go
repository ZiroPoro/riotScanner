package clone

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/edu/riotscanner/internal/adb"
)

type Provider string

const (
	ProviderADB      Provider = "adb"
	ProviderLDPlayer Provider = "ldplayer"
	ProviderNox      Provider = "nox"
)

type Config struct {
	Provider       Provider `yaml:"provider"`
	SourcePackage  string   `yaml:"source_package"`
	ClonePackage   string   `yaml:"clone_package"`
	LDConsolePath  string   `yaml:"ldconsole_path"`
	NoxConsolePath string   `yaml:"nox_console_path"`
	CloneAPKPath   string   `yaml:"clone_apk_path"`
}

type Manager struct {
	cfg    Config
	client *adb.Client
}

func NewManager(cfg Config, client *adb.Client) *Manager {
	return &Manager{cfg: cfg, client: client}
}

func (m *Manager) CreateClone() error {
	switch m.cfg.Provider {
	case ProviderLDPlayer:
		return m.createLDPlayerClone()
	case ProviderNox:
		return m.createNoxClone()
	case ProviderADB, "":
		return m.createADBClone()
	default:
		return fmt.Errorf("unknown clone provider: %s", m.cfg.Provider)
	}
}

func (m *Manager) DeleteClone() error {
	pkg := m.clonePackage()
	if pkg == "" {
		return fmt.Errorf("clone_package is required")
	}

	switch m.cfg.Provider {
	case ProviderLDPlayer:
		return m.runLDConsole("uninstallapp", "--packagename", pkg)
	case ProviderNox:
		return m.client.Uninstall(pkg)
	default:
		return m.client.Uninstall(pkg)
	}
}

func (m *Manager) ActivateClone() error {
	pkg := m.clonePackage()
	if pkg == "" {
		return fmt.Errorf("clone_package is required")
	}
	return m.client.StartApp(pkg)
}

func (m *Manager) clonePackage() string {
	if m.cfg.ClonePackage != "" {
		return m.cfg.ClonePackage
	}
	return m.cfg.SourcePackage + ".clone"
}

func (m *Manager) createADBClone() error {
	if m.cfg.CloneAPKPath != "" {
		return m.client.Install(m.cfg.CloneAPKPath)
	}
	return fmt.Errorf("adb provider requires clone_apk_path")
}

func (m *Manager) createLDPlayerClone() error {
	if m.cfg.LDConsolePath == "" {
		return fmt.Errorf("ldconsole_path is required for ldplayer provider")
	}
	src := m.cfg.SourcePackage
	dst := m.clonePackage()
	return m.runLDConsole("copyapp", "--from", src, "--to", dst, "--name", dst)
}

func (m *Manager) createNoxClone() error {
	if m.cfg.CloneAPKPath != "" {
		return m.client.Install(m.cfg.CloneAPKPath)
	}
	return fmt.Errorf("nox provider requires clone_apk_path")
}

func (m *Manager) runLDConsole(args ...string) error {
	cmd := exec.Command(m.cfg.LDConsolePath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ldconsole %s: %w: %s", strings.Join(args, " "), err, out)
	}
	return nil
}

func (m *Manager) ListPackages(filter string) ([]string, error) {
	out, err := m.client.Shell("pm", "list", "packages", filter)
	if err != nil {
		return nil, err
	}
	var pkgs []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package:") {
			pkgs = append(pkgs, strings.TrimPrefix(line, "package:"))
		}
	}
	return pkgs, nil
}
