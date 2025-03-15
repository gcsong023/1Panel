package systemctl

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

var ServiceCmd string

func init() {
	detectSystemManager()
}
func detectSystemManager() {
	ServiceCmd = "service" // 默认值
	switch {
	case detectSystemd():
		ServiceCmd = "systemctl"
	case detectOpenRC():
		ServiceCmd = "rc-service"
	case detectSysVinit():
		ServiceCmd = "service"
	}
}

func detectSystemd() bool {
	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return true
	}
	_, err := exec.LookPath("systemctl")
	return err == nil
}

func detectOpenRC() bool {
	_, err := exec.LookPath("rc-service")
	return err == nil
}

func detectSysVinit() bool {
	_, err := exec.LookPath("service")
	return err == nil
}
func serviceSuffix(serviceName string) string {
	switch ServiceCmd {
	case "systemctl":
		return serviceName
	case "rc-service", "service":
		if strings.HasSuffix(serviceName, ".service") {
			trimmed := strings.TrimSuffix(serviceName, ".service")
			return trimmed + "d"
		}
		return serviceName
	default:
		return serviceName
	}
}

func RunCommand(op string, serviceName string) (string, error) {
	var args []string
	switch ServiceCmd {
	case "systemctl":
		args = []string{string(op), serviceSuffix(serviceName)}
	case "rc-service":
		args = []string{serviceSuffix(serviceName), string(op)}
	case "service":
		args = []string{serviceSuffix(serviceName), string(op)}
	default:
		return "", fmt.Errorf("unsupported service manager: %s", ServiceCmd)
	}

	cmd := exec.Command(ServiceCmd, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command failed: %v", err)
	}
	return string(output), nil
}

// func StartService(serviceName string) (string, error) {
// 	return RunCommand(Start, serviceName)
// }

// func StopService(serviceName string) (string, error) {
// 	return RunCommand(Stop, serviceName)
// }

func RestartService(serviceName string) (string, error) {
	return RunCommand("restart", serviceName)
}

// func ServiceStatus(serviceName string) (string, error) {
// 	return RunCommand(Status, serviceName)
// }

func RunSystemCtl(args ...string) (string, error) {
	cmd := exec.Command("systemctl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("failed to run command: %w", err)
	}
	return string(output), nil
}

func IsActive(serviceName string) (bool, error) {
	out, err := RunSystemCtl("is-active", serviceName)
	if err != nil {
		return false, err
	}
	return out == "active\n", nil
}

func IsEnable(serviceName string) (bool, error) {
	out, err := RunSystemCtl("is-enabled", serviceName)
	if err != nil {
		return false, err
	}
	return out == "enabled\n", nil
}

func IsExist(serviceName string) (bool, error) {
	out, err := RunSystemCtl("is-enabled", serviceName)
	if err != nil {
		if strings.Contains(out, "disabled") {
			return true, nil
		}
		return false, nil
	}
	return true, nil
}

func handlerErr(out string, err error) error {
	if err != nil {
		if out != "" {
			return errors.New(out)
		}
		return err
	}
	return nil
}

func ServiceRestart(serviceName string) error {
	out, err := RunSystemCtl("restart", serviceName)
	return handlerErr(out, err)
}

func Operate(operate, serviceName string) error {
	out, err := RunSystemCtl(operate, serviceName)
	return handlerErr(out, err)
}
