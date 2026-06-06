package adb

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Client struct {
	adbPath string
	serial  string
}

func NewClient(adbPath, serial string) (*Client, error) {
	if adbPath == "" {
		adbPath = "adb"
	}

	if serial == "" {
		out, err := exec.Command(adbPath, "devices").CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("list devices: %w", err)
		}
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "List of devices") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 2 && parts[1] == "device" {
				serial = parts[0]
				break
			}
		}
		if serial == "" {
			return nil, fmt.Errorf("no adb devices connected")
		}
	}

	state, err := exec.Command(adbPath, "-s", serial, "get-state").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("device %s not ready: %w: %s", serial, err, state)
	}

	return &Client{adbPath: adbPath, serial: serial}, nil
}

func (c *Client) Serial() string {
	return c.serial
}

func (c *Client) adbPrefix() []string {
	return []string{"-s", c.serial}
}

func (c *Client) Shell(args ...string) (string, error) {
	cmdArgs := append(c.adbPrefix(), "shell")
	cmdArgs = append(cmdArgs, args...)
	out, err := exec.Command(c.adbPath, cmdArgs...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("shell %v: %w: %s", args, err, out)
	}
	return strings.TrimSpace(string(out)), nil
}

func (c *Client) Tap(x, y int) error {
	_, err := c.Shell("input", "tap", fmt.Sprintf("%d", x), fmt.Sprintf("%d", y))
	return err
}

func (c *Client) Swipe(x1, y1, x2, y2, durationMs int) error {
	_, err := c.Shell(
		"input", "swipe",
		fmt.Sprintf("%d", x1), fmt.Sprintf("%d", y1),
		fmt.Sprintf("%d", x2), fmt.Sprintf("%d", y2),
		fmt.Sprintf("%d", durationMs),
	)
	return err
}

func (c *Client) Screenshot() ([]byte, error) {
	cmd := exec.Command(c.adbPath, append(c.adbPrefix(), "exec-out", "screencap", "-p")...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("screencap: %w", err)
	}
	return out, nil
}

func (c *Client) Push(local, remote string) error {
	cmd := exec.Command(c.adbPath, append(c.adbPrefix(), "push", local, remote)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("push %s -> %s: %w: %s", local, remote, err, out)
	}
	return nil
}

func (c *Client) Install(apkPath string) error {
	abs, err := filepath.Abs(apkPath)
	if err != nil {
		return err
	}
	cmd := exec.Command(c.adbPath, append(c.adbPrefix(), "install", "-r", abs)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("install %s: %w: %s", abs, err, out)
	}
	return nil
}

func (c *Client) Uninstall(packageName string) error {
	cmd := exec.Command(c.adbPath, append(c.adbPrefix(), "uninstall", packageName)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("uninstall %s: %w: %s", packageName, err, out)
	}
	return nil
}

func (c *Client) StartApp(packageName string) error {
	_, err := c.Shell("monkey", "-p", packageName, "-c", "android.intent.category.LAUNCHER", "1")
	return err
}

func (c *Client) ForceStop(packageName string) error {
	_, err := c.Shell("am", "force-stop", packageName)
	return err
}

func WaitForDevice(adbPath, serial string, timeout time.Duration) error {
	if adbPath == "" {
		adbPath = "adb"
	}
	args := []string{"wait-for-device"}
	if serial != "" {
		args = append([]string{"-s", serial}, args...)
	}
	cmd := exec.Command(adbPath, args...)
	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("wait-for-device timeout after %s", timeout)
	}
}
