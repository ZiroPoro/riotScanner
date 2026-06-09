package camera

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Injector struct {
	adbPath string
	serial  string
}

func NewInjector(adbPath, serial string) *Injector {
	if adbPath == "" {
		adbPath = "adb"
	}
	return &Injector{adbPath: adbPath, serial: serial}
}

func (i *Injector) InjectImage(hostImagePath string) error {
	abs, err := filepath.Abs(hostImagePath)
	if err != nil {
		return err
	}
	if _, err := os.Stat(abs); err != nil {
		return fmt.Errorf("image not found: %w", err)
	}

	normalized := filepath.ToSlash(abs)

	args := []string{"emu", "camera", "load", normalized}
	if i.serial != "" {
		args = append([]string{"-s", i.serial}, args...)
	}

	cmd := exec.Command(i.adbPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("camera inject: %w: %s", err, out)
	}
	return nil
}

func (i *Injector) EnableVirtualScene() error {
	args := []string{"emu", "camera", "set", "mode", "virtualscene"}
	if i.serial != "" {
		args = append([]string{"-s", i.serial}, args...)
	}
	cmd := exec.Command(i.adbPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("camera mode: %w: %s", err, out)
	}
	return nil
}

func (i *Injector) PushToDevice(hostImagePath, remotePath string) error {
	args := []string{"push", hostImagePath, remotePath}
	if i.serial != "" {
		args = append([]string{"-s", i.serial}, args...)
	}
	cmd := exec.Command(i.adbPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("push image: %w: %s", err, out)
	}
	return nil
}
