package keys

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/filesystem"
)

// KeyringBackendFromEnv determines the appropriate keyring backend based on the environment.
// For mock/playground environments, it returns the test backend.
// For custom environments, it prompts the user to select between available backends (test and os if not on Darwin).
// For other environments, it returns the OS backend if not on Darwin, otherwise test backend.
func KeyringBackendFromEnv(env string) consts.SupportedKeyringBackend {
	switch env {
	case "mock", "playground", "blumbus":
		return consts.SupportedKeyringBackends.Test
	case "custom":
		krBackends := []string{"test"}
		if runtime.GOOS != "darwin" {
			krBackends = append(krBackends, "os")
		}
		keyringBackend, _ := pterm.DefaultInteractiveSelect.WithDefaultText(
			"select the keyring backend you want to use",
		).WithOptions(krBackends).Show()
		return consts.SupportedKeyringBackend(keyringBackend)
	default:
		if runtime.GOOS != "darwin" {
			return consts.SupportedKeyringBackends.OS
		}
		return consts.SupportedKeyringBackends.Test
	}
}

// RunCmdBasedOnKeyringBackend executes the given command with different behavior based on the keyring backend.
// For OS keyring backend, it reads the passphrase from a file and handles the interactive prompts.
// For other backends, it executes the command directly.
// Returns the command output buffer and any error encountered.
func RunCmdBasedOnKeyringBackend(
	home, command string,
	args []string,
	kb consts.SupportedKeyringBackend,
) (*bytes.Buffer, error) {
	var out *bytes.Buffer
	var pswFileName consts.OsKeyringPwdFileName

	pswFileName, err := filesystem.GetOsKeyringPswFileName(command)
	if err != nil {
		return nil, err
	}

	if kb == consts.SupportedKeyringBackends.OS {
		fp := filepath.Join(home, string(pswFileName))
		psw, err := filesystem.ReadFromFile(fp)
		if err != nil {
			return nil, err
		}

		pr := map[string]string{
			"Enter keyring passphrase":    psw,
			"Re-enter keyring passphrase": psw,
		}
		out, err = bash.ExecuteCommandWithPrompts(
			command, args, pr,
		)
		if err != nil {
			return nil, err
		}
	} else {
		var err error

		cmd := exec.Command(command, args...)

		out, err = bash.ExecCommandWithStdout(cmd)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
