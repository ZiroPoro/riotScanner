package emulator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/edu/riotscanner/internal/adb"
)

type Config struct {
	EmulatorPath string        `yaml:"emulator_path"`
	AVDName      string        `yaml:"avd_name"`
	ADBPath      string        `yaml:"adb_path"`
	Serial       string        `yaml:"serial"`
	ExtraArgs    []string      `yaml:"extra_args"`
	BootTimeout  time.Duration `yaml:"boot_timeout"`
}

type Manager struct {
	cfg    Config
	cmd    *exec.Cmd
	serial string
}

func NewManager(cfg Config) *Manager {
	if cfg.BootTimeout == 0 {
		cfg.BootTimeout = 3 * time.Minute
	}
	if cfg.ExtraArgs == nil {
		cfg.ExtraArgs = []string{
			"-camera-back", "none",
			"-camera-front", "emulated",
			"-no-snapshot-load",
		}
	}
	return &Manager{cfg: cfg}
}

func (m *Manager) Start(ctx context.Context) (string, error) {
	if m.cfg.EmulatorPath == "" {
		return "", fmt.Errorf("emulator_path is required")
	}
	if m.cfg.AVDName == "" {
		return "", fmt.Errorf("avd_name is required")
	}

	args := append([]string{"-avd", m.cfg.AVDName}, m.cfg.ExtraArgs...)
	m.cmd = exec.CommandContext(ctx, m.cfg.EmulatorPath, args...)
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	if err := m.cmd.Start(); err != nil {
		return "", fmt.Errorf("start emulator: %w", err)
	}

	serial := m.cfg.Serial
	if serial == "" {
		serial = "emulator-5554"
	}
	m.serial = serial

	if err := adb.WaitForDevice(m.cfg.ADBPath, serial, m.cfg.BootTimeout); err != nil {
		return "", err
	}

	deadline := time.Now().Add(m.cfg.BootTimeout)
	for time.Now().Before(deadline) {
		client, err := adb.NewClient(m.cfg.ADBPath, serial)
		if err == nil {
			val, _ := client.Shell("getprop", "sys.boot_completed")
			if strings.TrimSpace(val) == "1" {
				return serial, nil
			}
		}
		time.Sleep(2 * time.Second)
	}

	return serial, fmt.Errorf("emulator boot timeout")
}

func (m *Manager) Stop() error {
	if m.cmd == nil || m.cmd.Process == nil {
		return nil
	}
	return m.cmd.Process.Kill()
}

func ResolveSDKPaths(sdkRoot string) (emulatorPath, adbPath string, err error) {
	if sdkRoot == "" {
		sdkRoot = os.Getenv("ANDROID_HOME")
	}
	if sdkRoot == "" {
		sdkRoot = os.Getenv("ANDROID_SDK_ROOT")
	}
	if sdkRoot == "" {
		return "", "", fmt.Errorf("ANDROID_HOME or ANDROID_SDK_ROOT not set")
	}

	emulatorPath = filepath.Join(sdkRoot, "emulator", "emulator.exe")
	adbPath = filepath.Join(sdkRoot, "platform-tools", "adb.exe")

	if _, err := os.Stat(emulatorPath); err != nil {
		return "", "", fmt.Errorf("emulator not found at %s", emulatorPath)
	}
	if _, err := os.Stat(adbPath); err != nil {
		return "", "", fmt.Errorf("adb not found at %s", adbPath)
	}
	return emulatorPath, adbPath, nil
}
