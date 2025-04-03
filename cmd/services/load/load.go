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

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/cronjobs"
	"github.com/dymensionxyz/roller/utils/config/scripts"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/roller"
)

type Service struct {
	Name        string
	LogFilePath string
}

type ServiceTemplateData struct {
	Name         string
	ExecPath     string
	UserName     string
	CustomRunCmd []string
	Home         string
}

func Cmd(services []string, module string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "Loads the different RollApp services on the local machine",
		Run: func(cmd *cobra.Command, args []string) {
			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollerData, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			if rollerData.DA.Backend == consts.Celestia {
				services = consts.RollappWithCelesSystemdServices
			}

			err = LoadServices(services, rollerData)
			if err != nil {
				pterm.Error.Println("failed to load services: ", err)
				return
			}

			if module == "rollapp" {
				err = scripts.CreateRollappStartup(home)
				if err != nil {
					pterm.Error.Println("failed to generate startup scripts:", err)
					return
				}
			}

			if module == "relayer" {
				schedule := "*/15 * * * *" // Run every hour
				command := fmt.Sprintf(
					"%s tx flush hub-rollapp --max-msgs 100 --home %s",
					consts.Executables.Relayer,
					filepath.Join(rollerData.Home, consts.ConfigDirName.Relayer),
				)

				err := cronjobs.Add(schedule, command)
				if err != nil {
					pterm.Error.Println("failed to add flush cronjob", err)
					return
				}
			}

			defer func() {
				pterm.Info.Println("next steps:")
				switch module {
				case "rng", "price":
					pterm.Info.Printf(
						"run %s to start %s on your local machine\n",
						pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
							Sprintf("roller oracle %s services start", module),
						strings.Join(services, ", "),
					)
				default:
					pterm.Info.Printf(
						"run %s to start %s on your local machine\n",
						pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
							Sprintf("roller %s services start", module),
						strings.Join(services, ", "),
					)
				}
			}()
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

func generateCustomLaunchctlServiceTemplate(
	serviceData ServiceTemplateData,
	pArgs any,
) (*bytes.Buffer, error) {
	var args []string
	switch v := pArgs.(type) {
	case []string:
		args = v
	case *exec.Cmd:
		args = v.Args
	default:
		return nil, fmt.Errorf("unsupported type for programArgs: %T", pArgs)
	}

	tmpl := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>xyz.dymension.roller.{{.Name}}</string>

    <key>ProgramArguments</key>
    <array>
        {{range .CustomRunCmd}}
        <string>{{.}}</string>
        {{end}}
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

	serviceData.CustomRunCmd = args

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
	var tmpl string

	// The difference between the templates are:
	// for rollapp, the memory limit is higher
	// for rng and price, the start command is different as it's a subcommand based on the oracle type

	switch serviceData.Name {
	case "rng", "price":
		tmpl = `[Unit]
Description=Roller {{.Name}} service
After=network.target

[Service]
Environment="PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ExecStart={{.ExecPath}} oracle {{.Name}} start --home {{.Home}}
Restart=on-failure
RestartSec=10
MemoryHigh=15%
MemoryMax=20%
User={{.UserName}}
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
`
	case "rollapp":
		tmpl = `[Unit]
Description=Roller {{.Name}} service
After=network.target

[Service]
Environment="PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
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
	default:
		tmpl = `[Unit]
Description=Roller {{.Name}} service
After=network.target

[Service]
Environment="PATH=/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
ExecStart={{.ExecPath}} {{.Name}} start
Restart=on-failure
RestartSec=10
MemoryHigh=15%
MemoryMax=20%
User={{.UserName}}
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
`
	}

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

func LoadMacOsServices(services []string, rollerData roller.RollappConfig) error {
	for _, service := range services {
		serviceData := ServiceTemplateData{
			Name:     service,
			ExecPath: consts.Executables.Roller,
			UserName: os.Getenv("USER"),
		}

		var tpl *bytes.Buffer
		var err error

		if service == "da-light-client" {
			damanager := datalayer.NewDAManager(
				rollerData.DA.Backend,
				rollerData.Home,
				rollerData.KeyringBackend,
				rollerData.NodeType,
			)
			c := damanager.GetStartDACmd()

			// during the development of ~v1.6.4 there was an issue running
			// the da light client inside a launchctl service using the
			// built-in roller's `roller da-light-client start` command, hence this workaround
			// @20241011
			tpl, err = generateCustomLaunchctlServiceTemplate(serviceData, c)
			if err != nil {
				pterm.Error.Println("failed to generate template", err)
				return err
			}
		} else {
			tpl, err = generateLaunchctlServiceTemplate(serviceData)
			if err != nil {
				pterm.Error.Println("failed to generate template", err)
				return err
			}
		}

		err = writeLaunchctlServiceFile(tpl, service)
		if err != nil {
			pterm.Error.Println("failed to write launchctl file", err)
			return err
		}
	}
	pterm.Success.Printf(
		"💈 Services %s been loaded successfully.\n",
		strings.Join(services, ", "),
	)

	return nil
}

func LoadLinuxServices(services []string, rollerData roller.RollappConfig) error {
	for _, service := range services {
		serviceData := ServiceTemplateData{
			Name:     service,
			ExecPath: consts.Executables.Roller,
			UserName: os.Getenv("USER"),
			Home:     rollerData.Home,
		}
		tpl, err := generateSystemdServiceTemplate(serviceData)
		errorhandling.PrettifyErrorIfExists(err)
		err = writeSystemdServiceFile(tpl, service)
		errorhandling.PrettifyErrorIfExists(err)
	}

	_, err := bash.ExecCommandWithStdout(
		exec.Command("sudo", "systemctl", "daemon-reload"),
	)
	if err != nil {
		pterm.Error.Println("failed to write launchctl file", err)
		return err
	}

	pterm.Success.Printf(
		"💈 Services %s been loaded successfully.\n",
		strings.Join(services, ", "),
	)

	return nil
}

func LoadServices(services []string, rollerData roller.RollappConfig) error {
	if runtime.GOOS == "darwin" {
		err := LoadMacOsServices(services, rollerData)
		if err != nil {
			return err
		}
	} else if runtime.GOOS == "linux" {
		err := LoadLinuxServices(services, rollerData)
		if err != nil {
			return err
		}
	} else {
		err := fmt.Errorf(
			"only linux and macos is currently supported",
		)
		return err
	}

	return nil
}
