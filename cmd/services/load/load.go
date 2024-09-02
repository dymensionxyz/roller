package load

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

type Service struct {
	Name        string
	LogFilePath string
}

type ServiceTemplateData struct {
	Name     string
	ExecPath string
	UserName string
}

func Cmd(services []string, module string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "Loads the different RollApp services on the local machine",
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS != "linux" {
				pterm.Error.Printf(
					"the %s commands are only available on linux machines\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("'services'"),
				)

				return
			}
			for _, service := range services {
				serviceData := ServiceTemplateData{
					Name:     service,
					ExecPath: consts.Executables.Roller,
					UserName: os.Getenv("USER"),
				}
				tpl, err := generateServiceTemplate(serviceData)
				errorhandling.PrettifyErrorIfExists(err)
				err = writeServiceFile(tpl, service)
				errorhandling.PrettifyErrorIfExists(err)
			}
			_, err := bash.ExecCommandWithStdout(
				exec.Command("sudo", "systemctl", "daemon-reload"),
			)
			errorhandling.PrettifyErrorIfExists(err)
			pterm.Success.Printf(
				"💈 Services %s been loaded successfully.\n",
				strings.Join(services, ", "),
			)

			pterm.Info.Println("next steps:")
			pterm.Info.Printf(
				"run %s to start %s on your local machine\n",
				pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
					Sprintf("roller %s services start", module),
				strings.Join(services, ", "),
			)
		},
	}
	return cmd
}

func writeServiceFile(serviceTxt *bytes.Buffer, serviceName string) error {
	filePath := filepath.Join("/etc/systemd/system/", fmt.Sprintf("%s.service", serviceName))
	cmd := exec.Command(
		"bash", "-c", fmt.Sprintf(
			"echo '%s' | sudo tee %s",
			serviceTxt.String(), filePath,
		),
	)
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
