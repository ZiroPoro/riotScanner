package orchestrator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/edu/riotscanner/internal/adb"
	"github.com/edu/riotscanner/internal/automation"
	"github.com/edu/riotscanner/internal/camera"
	"github.com/edu/riotscanner/internal/clone"
	"github.com/edu/riotscanner/internal/config"
	"github.com/edu/riotscanner/internal/emulator"
)

type Orchestrator struct {
	cfg      *config.Config
	adb      *adb.Client
	emulator *emulator.Manager
	camera   *camera.Injector
	clone    *clone.Manager
	runner   *automation.Runner
	log      *log.Logger
}

func New(cfg *config.Config, logger *log.Logger) *Orchestrator {
	if logger == nil {
		logger = log.Default()
	}
	return &Orchestrator{
		cfg:      cfg,
		emulator: emulator.NewManager(cfg.Emulator),
		log:      logger,
	}
}

func (o *Orchestrator) StartEmulator(ctx context.Context) error {
	serial, err := o.emulator.Start(ctx)
	if err != nil {
		return err
	}
	o.cfg.Emulator.Serial = serial
	return o.connectADB()
}

func (o *Orchestrator) ConnectADB() error {
	return o.connectADB()
}

func (o *Orchestrator) connectADB() error {
	client, err := adb.NewClient(o.cfg.Emulator.ADBPath, o.cfg.Emulator.Serial)
	if err != nil {
		return err
	}
	o.adb = client
	o.camera = camera.NewInjector(o.cfg.Emulator.ADBPath, o.cfg.Emulator.Serial)
	o.clone = clone.NewManager(o.cfg.Clone, client)
	o.runner = automation.NewRunner(client, o.cfg.Automation, o.cfg.PollMs)
	return nil
}

func (o *Orchestrator) CreateClone() error {
	if o.clone == nil {
		return fmt.Errorf("not connected to device")
	}
	o.log.Println("creating app clone...")
	return o.clone.CreateClone()
}

func (o *Orchestrator) DeleteClone() error {
	if o.clone == nil {
		return fmt.Errorf("not connected to device")
	}
	o.log.Println("deleting app clone...")
	return o.clone.DeleteClone()
}

func (o *Orchestrator) InjectQR() error {
	if o.camera == nil {
		return fmt.Errorf("not connected to device")
	}
	path := o.cfg.QRScan.QRImagePath
	if path == "" {
		return fmt.Errorf("qr_image_path not configured")
	}
	o.log.Printf("injecting QR image: %s\n", path)
	return o.camera.InjectImage(path)
}

func (o *Orchestrator) RunAutomationLoop(ctx context.Context) error {
	if o.runner == nil {
		return fmt.Errorf("not connected to device")
	}

	qrMatcher := automation.NewScreenMatcher(o.cfg.QRScan.DetectProbes)
	successMatcher := automation.NewScreenMatcher(o.cfg.QRScan.SuccessProbes)
	qrInjected := false

	o.log.Println("starting automation loop...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		raw, err := o.adb.Screenshot()
		if err != nil {
			o.log.Printf("screenshot error: %v\n", err)
			time.Sleep(time.Duration(o.cfg.PollMs) * time.Millisecond)
			continue
		}

		img, err := automation.LoadImage(raw)
		if err != nil {
			o.log.Printf("decode error: %v\n", err)
			continue
		}

		if o.cfg.QRScan.Enabled && !qrInjected && qrMatcher.MatchAll(img) {
			o.log.Println("QR scan screen detected, injecting camera image...")
			if err := o.InjectQR(); err != nil {
				o.log.Printf("camera inject error: %v\n", err)
			} else {
				qrInjected = true
				wait := o.cfg.QRScan.WaitAfterMs
				if wait > 0 {
					time.Sleep(time.Duration(wait) * time.Millisecond)
				}
			}
		}

		if o.cfg.QRScan.Enabled && qrInjected && successMatcher.MatchAll(img) {
			o.log.Println("QR scan success detected")
			qrInjected = false
		}

		for _, action := range o.cfg.Automation {
			matcher := automation.NewScreenMatcher(action.Probes)
			if matcher.MatchAll(img) {
				o.log.Printf("matched action: %s -> tap (%d,%d)\n", action.Name, action.TapX, action.TapY)
				if err := o.adb.Tap(action.TapX, action.TapY); err != nil {
					o.log.Printf("tap error: %v\n", err)
				}
				break
			}
		}

		time.Sleep(time.Duration(o.cfg.PollMs) * time.Millisecond)
	}
}

func (o *Orchestrator) Stop() error {
	return o.emulator.Stop()
}
