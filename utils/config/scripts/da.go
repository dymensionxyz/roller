package scripts

import (
	"bytes"
	"html/template"
)

func generateDaStartupScript(d DaStartupTemplateData) (*bytes.Buffer, error) {
	tmpl := `#!/usr/bin/expect -f

# Set the timeout to a large number to continuously wait for the prompt
set timeout -1

# AWS Secret Manager Example
# set password [exec aws secretsmanager get-secret-value --secret-id {{ secret_path }} --query SecretString --output text --no-cli-pager]
set password [cat {{.PasswordFilePath}}]

spawn {{.Binary}} light start --core.ip {{.StateNode}} --node.store {{HomeDir}} --keyring.backend {{.KeyringBackend}}

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
