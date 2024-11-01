package scripts

import (
	"bytes"
	"html/template"
)

func generateRollappStartupScript(d RaStartupTemplateData) (*bytes.Buffer, error) {
	tmpl := `#!/usr/bin/expect -f

# Set the timeout to a large number to continuously wait for the prompt
set timeout -1

# AWS Secret Manager Example
# set password [exec aws secretsmanager get-secret-value --secret-id {{ secret_path }} --query SecretString --output text --no-cli-pager]
set password [cat {{.PasswordFilePath}}]

spawn {{.Binary}} start --home {{.HomeDir}} --keyring-backend {{.KeyringBackend}}

# Loop waiting for the password prompt
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

# End the script
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
