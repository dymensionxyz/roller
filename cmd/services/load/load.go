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

			if runtime.GOOS == "darwin" {
				for _, service := range services {
					serviceData := ServiceTemplateData{
						Name:     service,
						ExecPath: consts.Executables.Roller,
						UserName: os.Getenv("USER"),
					}
					tpl, err := generateLaunchctlServiceTemplate(serviceData)
					if err != nil {
						pterm.Error.Println("failed to generate template", err)
						return
					}
					err = writeLaunchctlServiceFile(tpl, service)
					if err != nil {
						pterm.Error.Println("failed to write launchctl file", err)
						return
					}
					errorhandling.PrettifyErrorIfExists(err)
					filePath := filepath.Join(
						"/Library/LaunchDaemons/",
						fmt.Sprintf("xyz.dymension.roller.%s.plist", service),
					)

					_, err = bash.ExecCommandWithStdout(
						exec.Command(
							"sudo",
							"launchctl",
							"load",
							filePath,
						),
					)

					errorhandling.PrettifyErrorIfExists(err)
				}

				return
			} else if runtime.GOOS == "linux" {
				for _, service := range services {
					serviceData := ServiceTemplateData{
						Name:     service,
						ExecPath: consts.Executables.Roller,
						UserName: os.Getenv("USER"),
					}
					tpl, err := generateSystemdServiceTemplate(serviceData)
					errorhandling.PrettifyErrorIfExists(err)
					err = writeSystemdServiceFile(tpl, service)
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

			} else {
				pterm.Info.Printf(
					"the %s commands currently support only darwin and linux operating systems",
					cmd.Use,
				)
				return
			}

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

// TODO: refactor into generic service functions that handle different operating systems
func writeLaunchctlServiceFile(serviceTxt *bytes.Buffer, serviceName string) error {
	filePath := filepath.Join(
		"/Library/LaunchDaemons/",
		fmt.Sprintf("xyz.dymension.roller.%s.plist", serviceName),
	)
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

func writeSystemdServiceFile(serviceTxt *bytes.Buffer, serviceName string) error {
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

func generateLaunchctlServiceTemplate(
	serviceData ServiceTemplateData,
) (*bytes.Buffer, error) {
	tmpl := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>xyz.dymension.roller.{{.Name}}</string>

    <key>ProgramArguments</key>
    <array>
        <string>{{.ExecPath}}</string>
        <string>{{.Name}}</string>
        <string>start</string>
    </array>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
    </dict>

    <key>ThrottleInterval</key>
    <integer>10</integer>

    <key>UserName</key>
    <string>{{.UserName}}</string>

    <key>SoftResourceLimits</key>
    <dict>
        <key>NumberOfFiles</key>
        <integer>65535</integer>
    </dict>

    <key>HardResourceLimits</key>
    <dict>
        <key>NumberOfFiles</key>
        <integer>65535</integer>
    </dict>

    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin/roller_bins:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
    </dict>
</dict>
</plist>
`
	serviceTemplate, err := template.New("service").Parse(tmpl)
	if err != nil {
		pterm.Println("failed to create template")
		return nil, err
	}
	var tpl bytes.Buffer
	err = serviceTemplate.Execute(&tpl, serviceData)
	if err != nil {
		pterm.Println("failed to generate template")
		return nil, err
	}
	return &tpl, nil
}

func generateSystemdServiceTemplate(serviceData ServiceTemplateData) (*bytes.Buffer, error) {
	tmpl := `[Unit]
Description=Roller {{.Name}} service
After=network.target

[Service]
ExecStart={{.ExecPath}} {{.Name}} start
Restart=on-failure
RestartSec=10
MemoryHigh=65%
MemoryMax=70%
User={{.UserName}}
LimitNOFILE=65535

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
