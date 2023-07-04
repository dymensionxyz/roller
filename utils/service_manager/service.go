package servicemanager

import (
	"context"
	"log"
	"math/big"
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
	Services  map[string]ServiceData
}

type UIData struct {
	//TODO: try to remove as it stored in a map
	Name     string
	Accounts []utils.AccountData
	Balance  string
	Status   string
}

type ServiceData struct {
	Command *exec.Cmd
	FetchFn func(config.RollappConfig) (*utils.AccountData, error)
	UIData  UIData
}

// TODO: move this to a separate file
func activeIfSufficientBalance(currentBalance, threshold *big.Int) string {
	if currentBalance.Cmp(threshold) >= 0 {
		return "Active"
	} else {
		return "Stopped"
	}
}

// TODO: fetch all data and populate UIData
func (s *ServiceConfig) FetchServicesData(cfg config.RollappConfig) {
	for k, service := range s.Services {
		//TODO: make this async
		if service.FetchFn != nil {
			accountData, err := service.FetchFn(cfg)
			if err != nil {
				//TODO: set the status to FAILED
				return
			}
			service.UIData.Accounts = []utils.AccountData{*accountData}

			//FIXME: fix the denom
			service.UIData.Balance = accountData.Balance.String()

			//FIXME: the status function should be part of the service
			service.UIData.Status = activeIfSufficientBalance(accountData.Balance, big.NewInt(1))
			if k == "Relayer" {
				service.UIData.Status = "Starting..."
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

func (s *ServiceConfig) AddService(name string, data ServiceData) {
	if s.Services == nil {
		s.Services = make(map[string]ServiceData)
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
