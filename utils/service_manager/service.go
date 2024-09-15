package servicemanager

import (
	"context"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
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
	Accounts []utils.AccountData
	Balance  string
	Status   string
}

// Service TODO: The relayer, sequencer and data layer should implement the Service interface (#208)
type Service struct {
	Command  *exec.Cmd
	FetchFn  func(config.RollappConfig) ([]utils.AccountData, error)
	StatusFn func(config.RollappConfig) string
	UIData   UIData
}

// TODO: fetch all data and populate UIData
func (s *ServiceConfig) FetchServicesData(cfg config.RollappConfig) {
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

func (s *ServiceConfig) InitServicesData(cfg config.RollappConfig) {
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

func StartSystemdService(serviceName string, options ...string) error {
	// Create the base command
	cmd := []string{"sudo", "systemctl", "start", serviceName}

	// Append any additional options
	cmd = append(cmd, options...)

	// Create the exec.Command
	execCmd := exec.Command(cmd[0], cmd[1:]...)

	// Execute the command
	err := bash.ExecCmd(execCmd)
	if err != nil {
		return err
	}

	return nil
}

func RestartSystemdService(serviceName string) error {
	// not ideal, shouldn't run sudo commands from within roller
	cmd := exec.Command("sudo", "systemctl", "restart", serviceName)
	err := bash.ExecCmd(cmd)
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
