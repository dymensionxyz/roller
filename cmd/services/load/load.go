package load

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

type Service struct {
	Name        string
	LogFilePath string
}

type ServiceTemplateData struct {
	Name        string
	ExecPath    string
	UserName    string
	LogFilePath string
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "Loads the different rollapp services on the local machine",
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS != "linux" {
				utils.PrettifyErrorIfExists(errors.New("the services commands are only available on linux machines"))
			}
			rollerHome := cmd.Flag(utils.FlagNames.Home).Value.String()
			services := getRollappServices(rollerHome)
			for _, service := range services {
				serviceData := ServiceTemplateData{
					Name:        service.Name,
					ExecPath:    consts.Executables.Roller,
					UserName:    os.Getenv("USER"),
					LogFilePath: service.LogFilePath,
				}
				tpl, err := generateServiceTemplate(serviceData)
				utils.PrettifyErrorIfExists(err)
				err = writeServiceFile(tpl, service.Name)
				utils.PrettifyErrorIfExists(err)
			}
			_, err := utils.ExecBashCommandWithStdout(exec.Command("sudo", "systemctl", "daemon-reload"))
			utils.PrettifyErrorIfExists(err)
			fmt.Println("💈 Services 'sequencer', 'da-light-client' and 'relayer' been loaded successfully." +
				" To start them, use 'systemctl start <service>'.")
		},
	}
	return cmd
}

func writeServiceFile(serviceTxt *bytes.Buffer, serviceName string) error {
	filePath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	cmd := exec.Command("bash", "-c", fmt.Sprintf("echo '%s' | sudo tee %s",
		serviceTxt.String(), filePath))
	// Need to start and wait instead of run to allow sudo to prompt for password
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func generateServiceTemplate(serviceData ServiceTemplateData) (*bytes.Buffer, error) {
	tmpl := `[Unit]
Description=Roller {{.Name}} service

[Service]
ExecStart={{.ExecPath}} {{.Name}} start
StandardOutput=file:{{.LogFilePath}}
StandardError=file:{{.LogFilePath}}
Restart=always
RestartSec=3s
User={{.UserName}}

[Install]
WantedBy=multi-user.target
`
	serviceTemplate, err := template.New("service").Parse(tmpl)
	if err != nil {
		return nil, err
	}
	var tpl bytes.Buffer
	err = serviceTemplate.Execute(&tpl, serviceData)
	if err != nil {
		return nil, err
	}
	return &tpl, nil
}

func getRollappServices(rollerHome string) []Service {
	return []Service{
		{
			Name:        "sequencer",
			LogFilePath: filepath.Join(rollerHome, consts.ConfigDirName.Rollapp, "rollapp.log"),
		},
		{
			Name:        "da-light-client",
			LogFilePath: filepath.Join(rollerHome, consts.ConfigDirName.DALightNode, "light_client.log"),
		},
		{
			Name:        "relayer",
			LogFilePath: filepath.Join(rollerHome, consts.ConfigDirName.Relayer, "relayer.log"),
		},
	}
}
