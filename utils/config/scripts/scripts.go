package scripts

import (
	"bytes"
	"fmt"
	"html/template"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func writeTemplateToFile(tmpl *bytes.Buffer, fp string) error {
	cmd := exec.Command(
		"bash", "-c", fmt.Sprintf(
			"echo '%s' | sudo tee %s",
			tmpl.String(), fp,
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

func writeDaTemplateToFile(home string) error {
	pswFp := filepath.Join(home, string(consts.OsKeyringPwdFileNames.Da))
	scriptFp := filepath.Join(home, string(consts.StartupScriptFilePaths.Da))

	daD := StartupTemplateData{
		PasswordFilePath: pswFp,
		Component:        "da-light-client",
		Home:             home,
	}

	daTmpl, err := generateStartupScript(daD)
	if err != nil {
		return fmt.Errorf("failed to generate DA startup script: %w", err)
	}

	err = writeTemplateToFile(daTmpl, scriptFp)
	if err != nil {
		return fmt.Errorf("failed to write DA startup script: %w", err)
	}

	return nil
}

func writeRaTemplateToFile(home string) error {
	pswFp := filepath.Join(home, string(consts.OsKeyringPwdFileNames.RollApp))
	scriptFp := filepath.Join(home, string(consts.StartupScriptFilePaths.RollApp))

	raD := StartupTemplateData{
		PasswordFilePath: pswFp,
		Component:        "rollapp",
		Home:             home,
	}

	raTmpl, err := generateStartupScript(raD)
	if err != nil {
		return err
	}

	err = writeTemplateToFile(raTmpl, scriptFp)
	if err != nil {
		return err
	}

	return nil
}

func CreateRollappStartup(home string) error {
	err := writeDaTemplateToFile(home)
	if err != nil {
		return fmt.Errorf("failed to generate DA startup script: %w", err)
	}

	err = writeRaTemplateToFile(home)
	if err != nil {
		return fmt.Errorf("failed to generate RollApp startup script: %w", err)
	}

	return nil
}

func generateStartupScript(d StartupTemplateData) (*bytes.Buffer, error) {
	tmpl := `#!/usr/bin/expect -f

# Set the timeout to a large number to continuously wait for the prompt
set timeout -1

# AWS Secret Manager Example
# set password [exec aws secretsmanager get-secret-value --secret-id <secret_path> --query SecretString --output text --no-cli-pager]

set password [exec cat {{.PasswordFilePath}}]

spawn roller {{.Component}} start --home {{.Home}}

expect {
    "Enter keyring passphrase:" {
        log_user 0
        send "$password\r"
        log_user 1
        exp_continue
    }
    "Re-enter keyring passphrase:" {
        log_user 0
        send "$password\r"
        log_user 1
        exp_continue
    }
    eof {
        break
    }
}

expect eof
`

	t, err := template.New("script").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var resp bytes.Buffer
	err = t.Execute(&resp, d)
	if err != nil {
		return nil, err
	}

	return &resp, err
}
