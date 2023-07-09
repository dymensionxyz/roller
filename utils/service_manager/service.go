package servicemanager

import (
	"context"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

type ServiceConfig struct {
	Context   context.Context
	WaitGroup *sync.WaitGroup
	Logger    *log.Logger
	Services  map[string]Service
}

type UIData struct {
	//TODO: try to remove as it stored in a map
	Name     string
	Accounts []utils.AccountData
	Balance  string
	Status   string
}

// Service TODO: The relayer, sequencer and data layer should implement the Service interface (#208)
type Service struct {
	Command *exec.Cmd

	// The fetch function should return the accounts data for the service. It should always return a list of accounts
	// in the same length, even if some of the fetches failed. In that case the account balance should be nil.
	FetchFn  func(config.RollappConfig) ([]utils.AccountData, error)
	StatusFn func(config.RollappConfig) string
	UIData   UIData
}

func (s *ServiceConfig) FetchServicesData(cfg config.RollappConfig) {
	for k, service := range s.Services {
		if service.FetchFn != nil {
			accountsData, err := service.FetchFn(cfg)
			if err != nil {
				logger := utils.GetRollerLogger(cfg.Home)
				logger.Printf("Error: failed to fetch data for service %s: %s", k, err.Error())
			}
			service.UIData.Accounts = accountsData
			if service.StatusFn != nil {
				service.UIData.Status = service.StatusFn(cfg)
			}
			s.Services[k] = service
		}
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

// FIXME(#154): this functions have busy loop in case some process fails to start
func (s *ServiceConfig) RunServiceWithRestart(name string, options ...utils.CommandOption) {
	if _, ok := s.Services[name]; !ok {
		panic("service with that name does not exist")
	}
	cmd := s.Services[name].Command

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
				time.Sleep(5 * time.Second)
				continue
			}
		}
	}()
}
