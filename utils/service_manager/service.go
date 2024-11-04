package servicemanager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

type ServiceConfig struct {
	Context   context.Context
	WaitGroup *sync.WaitGroup
	Logger    *log.Logger
	Services  map[string]Service
}

type UIData struct {
	// TODO: try to remove as it stored in a map
	Name     string
	Accounts []keys.AccountData
	Balance  string
	Status   string
}

// Service TODO: The relayer, sequencer and data layer should implement the Service interface (#208)
type Service struct {
	Command  *exec.Cmd
	FetchFn  func(roller.RollappConfig) ([]keys.AccountData, error)
	StatusFn func(roller.RollappConfig) string
	UIData   UIData
}

// TODO: fetch all data and populate UIData
func (s *ServiceConfig) FetchServicesData(cfg roller.RollappConfig) {
	for k, service := range s.Services {
		if service.FetchFn != nil {
			accountData, err := service.FetchFn(cfg)
			if err != nil {
				s.Logger.Println(err)
				continue
			}
			service.UIData.Accounts = accountData
			if service.StatusFn != nil {
				service.UIData.Status = service.StatusFn(cfg)
			}

			s.Services[k] = service
		}
	}
}

func (s *ServiceConfig) InitServicesData(cfg roller.RollappConfig) {
	for k, service := range s.Services {
		service.UIData.Status = "Starting..."
		s.Services[k] = service
	}
}

func (s *ServiceConfig) GetUIData() []UIData {
	var uiData []UIData
	for _, service := range s.Services {
		uiData = append(uiData, service.UIData)
	}
	return uiData
}

func (s *ServiceConfig) AddService(name string, data Service) {
	if s.Services == nil {
		s.Services = make(map[string]Service)
	}

	s.Services[name] = data
}

func (s *ServiceConfig) RunServiceWithRestart(name string, options ...bash.CommandOption) {
	if _, ok := s.Services[name]; !ok {
		panic("service with that name does not exist")
	}
	cmd := s.Services[name].Command
	if cmd == nil {
		s.Logger.Printf("service %s does not need to run separately", name)
		return
	}

	s.WaitGroup.Add(1)
	go func() {
		defer s.WaitGroup.Done()
		for {
			newCmd := exec.CommandContext(s.Context, cmd.Path, cmd.Args[1:]...)
			for _, option := range options {
				option(newCmd)
			}
			commandExited := make(chan error, 1)
			go func() {
				s.Logger.Printf("starting service command %s", newCmd.String())
				commandExited <- newCmd.Run()
			}()
			select {
			case <-s.Context.Done():
				return
			case <-commandExited:
				s.Logger.Printf("process %s exited, restarting...", newCmd.String())
				// FIXME(#154): this functions have busy loop in case some process fails to start
				time.Sleep(5 * time.Second)
				continue
			}
		}
	}()
}

func StartSystemdService(serviceName string) error {
	cmd := exec.Command("sudo", "systemctl", "start", serviceName)

	err := bash.ExecCmd(cmd)
	if err != nil {
		return err
	}

	return nil
}

func StartLaunchctlService(serviceName string) error {
	svcFilaPath := fmt.Sprintf("/Library/LaunchDaemons/xyz.dymension.roller.%s.plist", serviceName)

	err := exec.Command("sudo", "launchctl", "unload", "-w", svcFilaPath).Run()
	if err != nil {
		return err
	}

	err = exec.Command("sudo", "launchctl", "load", "-w", svcFilaPath).Run()
	if err != nil {
		return err
	}

	return nil
}

func RestartSystemdService(serviceName string) error {
	cmd := exec.Command("sudo", "systemctl", "restart", serviceName)

	// not ideal, shouldn't run sudo commands from within roller
	err := bash.ExecCmd(cmd)
	if err != nil {
		return err
	}

	return nil
}

func RestartLaunchctlService(serviceName string) error {
	svcFilaPath := fmt.Sprintf("/Library/LaunchDaemons/xyz.dymension.roller.%s.plist", serviceName)
	dCmd := exec.Command("sudo", "launchctl", "unload", "-w", svcFilaPath)
	err := bash.ExecCmd(dCmd)
	if err != nil {
		return err
	}

	uCmd := exec.Command("sudo", "launchctl", "load", "-w", svcFilaPath)
	err = bash.ExecCmd(uCmd)
	if err != nil {
		return err
	}

	return nil
}

func StopSystemdService(serviceName string) error {
	// not ideal, shouldn't run sudo commands from within roller
	cmd := exec.Command("sudo", "systemctl", "stop", serviceName)
	err := bash.ExecCmd(cmd)
	if err != nil {
		return err
	}

	return nil
}

func StopLaunchdService(serviceName string) error {
	svcFilaPath := fmt.Sprintf("/Library/LaunchDaemons/xyz.dymension.roller.%s.plist", serviceName)
	cmd := exec.Command(
		"sudo",
		"launchctl",
		"unload",
		"-w",
		svcFilaPath,
	)
	err := bash.ExecCmd(cmd)
	if err != nil {
		return err
	}

	return nil
}

func StopSystemServices() error {
	pterm.Info.Println("stopping existing system services, if any...")
	switch runtime.GOOS {
	case "linux":
		for _, svc := range consts.RollappSystemdServices {
			err := StopSystemdService(svc)
			if err != nil {
				pterm.Error.Println("failed to stop systemd service: ", err)
				return err
			}
		}
	case "darwin":
		for _, svc := range consts.RollappSystemdServices {
			err := StopLaunchdService(svc)
			if err != nil {
				pterm.Error.Println("failed to remove systemd service: ", err)
				return err
			}
		}
	default:
		pterm.Error.Println("OS not supported")
		return errors.New("OS not supported")
	}

	return nil
}
