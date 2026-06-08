package config

import (
	"fmt"
	"os"
	"time"

	"github.com/edu/riotscanner/internal/automation"
	"github.com/edu/riotscanner/internal/clone"
	"github.com/edu/riotscanner/internal/emulator"
	"gopkg.in/yaml.v3"
)

type QRScanConfig struct {
	Enabled       bool                    `yaml:"enabled"`
	DetectProbes  []automation.PixelProbe `yaml:"detect_probes"`
	QRImagePath   string                  `yaml:"qr_image_path"`
	WaitAfterMs   int                     `yaml:"wait_after_ms"`
	SuccessProbes []automation.PixelProbe `yaml:"success_probes"`
}

type Config struct {
	Emulator   emulator.Config        `yaml:"emulator"`
	Clone      clone.Config           `yaml:"clone"`
	Automation []automation.TapAction `yaml:"automation"`
	QRScan     QRScanConfig           `yaml:"qr_scan"`
	PollMs     int                    `yaml:"poll_ms"`
	LogLevel   string                 `yaml:"log_level"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Emulator.BootTimeout == 0 {
		cfg.Emulator.BootTimeout = 3 * time.Minute
	}
	if cfg.PollMs == 0 {
		cfg.PollMs = 500
	}

	return &cfg, nil
}
